package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerStorageAddons(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "s3-storage",
		Name:  "S3 Storage",
		Group: core.GroupStorage,
		Provides: []core.Capability{
			"storage:selected",
			"storage:s3",
			"provider:aws-s3",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return storageFragment(
					selection,
					"s3-storage",
					"S3",
					[]string{
						"AWS_REGION=us-east-1",
						"S3_BUCKET=replace-me",
						"AWS_ACCESS_KEY_ID=replace-me",
						"AWS_SECRET_ACCESS_KEY=replace-me",
					},
					"s3",
					jsS3StorageSource(),
					goS3StorageSource(),
				)
			},
		},
		Properties: map[string]string{"kind": "addon"},
	})

	r.RegisterEntry(EntryDefinition{
		ID:    "r2-storage",
		Name:  "Cloudflare R2 Storage",
		Group: core.GroupStorage,
		Provides: []core.Capability{
			"storage:selected",
			"storage:r2",
			"provider:cloudflare-r2",
		},
		RequiresAll: []core.Capability{"backend:selected"},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				return storageFragment(
					selection,
					"r2-storage",
					"Cloudflare R2",
					[]string{
						"R2_ACCOUNT_ID=replace-me",
						"R2_BUCKET=replace-me",
						"R2_ACCESS_KEY_ID=replace-me",
						"R2_SECRET_ACCESS_KEY=replace-me",
					},
					"r2",
					jsR2StorageSource(),
					goR2StorageSource(),
				)
			},
		},
		Properties: map[string]string{
			"kind":        "addon",
			"skillSource": "https://github.com/cloudflare/skills",
		},
	})
}

func storageFragment(selection core.TemplateSelection, ownerID string, providerName string, envLines []string, fileStem string, jsSource string, goSource string) []core.PlanFragment {
	actions := make([]core.ExecutionAction, 0, 3)
	if selection.Single(core.GroupBackend) != "hono-cf-workers" || fileStem != "r2" {
		actions = append(actions, writeEnvAction(
			fileStem+"-storage-env",
			"Write "+providerName+" storage environment template",
			"Adds "+providerName+" storage credentials to the backend environment template.",
			filepath.Join(backendDir(selection), ".env.example"),
			envLines,
		))
	}

	switch selection.Single(core.GroupBackend) {
	case "go-api":
		actions = append(actions, writeFileAction(fileStem+"-storage-go", "Write "+providerName+" storage helper", "Adds a starter "+providerName+" storage client for the Go backend.", filepath.Join(backendDir(selection), "internal", "storage", fileStem+".go"), goSource))
		actions = append(actions, commandAction(fileStem+"-storage-go-sdk", "Install "+providerName+" storage SDK", "Adds AWS SDK packages used by the Go storage helper.", backendDir(selection), "go", "get", "github.com/aws/aws-sdk-go-v2/config@latest", "github.com/aws/aws-sdk-go-v2/service/s3@latest"))
	case "hono-cf-workers":
		if fileStem == "r2" {
			actions = append(actions, writeFileAction("r2-storage-worker", "Write Cloudflare R2 binding helper", "Adds a native R2 helper for the Worker binding.", filepath.Join(backendDir(selection), "src", "storage", "r2.ts"), workerR2StorageSource()))
			break
		}
		actions = append(actions, writeFileAction(fileStem+"-storage-js", "Write "+providerName+" storage helper", "Adds a starter "+providerName+" storage client for the selected JS backend.", filepath.Join(backendDir(selection), "src", "storage", fileStem+".ts"), jsSource))
		actions = append(actions, packageManagerInstallActions(selection, packageManager(selection), backendDir(selection), "Install "+providerName+" storage SDK", "Adds the AWS S3 client used by the storage helper.", []string{"@aws-sdk/client-s3"}, nil)...)
	default:
		actions = append(actions, writeFileAction(fileStem+"-storage-js", "Write "+providerName+" storage helper", "Adds a starter "+providerName+" storage client for the selected JS backend.", filepath.Join(backendDir(selection), "src", "storage", fileStem+".ts"), jsSource))
		actions = append(actions, packageManagerInstallActions(selection, packageManager(selection), backendDir(selection), "Install "+providerName+" storage SDK", "Adds the AWS S3 client used by the storage helper.", []string{"@aws-sdk/client-s3"}, nil)...)
	}

	return []core.PlanFragment{{
		ID:      ownerID + "-addon",
		OwnerID: ownerID,
		Phase:   core.PhaseIntegration,
		Actions: actions,
	}}
}

func jsS3StorageSource() string {
	return `import { S3Client } from "@aws-sdk/client-s3";

export const storageClient = new S3Client({
  region: process.env.AWS_REGION,
});`
}

func jsR2StorageSource() string {
	return `import { S3Client } from "@aws-sdk/client-s3";

const accountId = process.env.R2_ACCOUNT_ID;

export const storageClient = new S3Client({
  region: "auto",
  endpoint: accountId ? ` + "`https://${accountId}.r2.cloudflarestorage.com`" + ` : undefined,
  credentials: {
    accessKeyId: process.env.R2_ACCESS_KEY_ID || "",
    secretAccessKey: process.env.R2_SECRET_ACCESS_KEY || "",
  },
});`
}

func workerR2StorageSource() string {
	return `export async function putObject(bucket: R2Bucket, key: string, value: ReadableStream | ArrayBuffer | string) {
  const object = await bucket.put(key, value);
  console.log("cloudflare r2 object written", { key });
  return object;
}

export async function getObject(bucket: R2Bucket, key: string) {
  const object = await bucket.get(key);
  console.log("cloudflare r2 object read", { key, found: object !== null });
  return object;
}`
}

func goS3StorageSource() string {
	return `package storage

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewS3Client(ctx context.Context) (*s3.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg), nil
}`
}

func goR2StorageSource() string {
	return `package storage

import (
	"context"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func NewR2Client(ctx context.Context) (*s3.Client, error) {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}
	return s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = awsString("https://" + accountID + ".r2.cloudflarestorage.com")
	}), nil
}

func awsString(value string) *string {
	return &value
}`
}
