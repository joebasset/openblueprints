package registry

import (
	"path/filepath"

	"openblueprints/internals/core"
)

func registerEmailAddons(r *Registry) {
	r.RegisterEntry(EntryDefinition{
		ID:    "resend",
		Name:  "Resend Email",
		Group: core.GroupAddon,
		Provides: []core.Capability{
			"addon:email",
			"provider:resend",
		},
		RequiresAll: []core.Capability{
			"backend:js",
			"runtime:js",
		},
		Fragments: []core.FragmentBuilder{
			func(selection core.TemplateSelection) []core.PlanFragment {
				actions := packageManagerInstallActions(selection, packageManager(selection), backendDir(selection), "Install Resend", "Adds the Resend SDK to the backend workspace.", []string{"resend"}, nil)
				actions = append(actions,
					writeEnvAction(
						"resend-env",
						"Write Resend environment template",
						"Adds Resend email credentials to the backend environment template.",
						filepath.Join(backendDir(selection), ".env.example"),
						[]string{
							"RESEND_API_KEY=re_replace_me",
							"RESEND_FROM_EMAIL=Acme <onboarding@example.com>",
						},
					),
					writeFileAction("resend-email-helper", "Write Resend email helper", "Adds a small helper for sending transactional email.", filepath.Join(backendDir(selection), "src", "email", "resend.ts"), resendEmailSource(selection)),
				)
				return []core.PlanFragment{{
					ID:      "resend-addon",
					OwnerID: "resend",
					Phase:   core.PhaseIntegration,
					Actions: actions,
				}}
			},
		},
		Properties: map[string]string{
			"kind":        "addon",
			"skillSource": "resend/resend-skills",
		},
	})
}

func resendEmailSource(selection core.TemplateSelection) string {
	if selection.Single(core.GroupBackend) == "hono-cf-workers" {
		return `import { Resend } from "resend";

type SendEmailInput = {
  apiKey: string;
  from: string;
  to: string;
  subject: string;
  html: string;
};

export async function sendEmail(input: SendEmailInput) {
  const resend = new Resend(input.apiKey);
  const result = await resend.emails.send({
    from: input.from,
    to: input.to,
    subject: input.subject,
    html: input.html,
  });
  console.log("resend email send requested", { to: input.to, subject: input.subject });
  return result;
}`
	}

	return `import { Resend } from "resend";

const apiKey = process.env.RESEND_API_KEY;
const fromEmail = process.env.RESEND_FROM_EMAIL;

if (!apiKey) {
  throw new Error("RESEND_API_KEY is required");
}

if (!fromEmail) {
  throw new Error("RESEND_FROM_EMAIL is required");
}

const resend = new Resend(apiKey);

type SendEmailInput = {
  to: string;
  subject: string;
  html: string;
};

export async function sendEmail(input: SendEmailInput) {
  const result = await resend.emails.send({
    from: fromEmail,
    to: input.to,
    subject: input.subject,
    html: input.html,
  });
  console.log("resend email send requested", { to: input.to, subject: input.subject });
  return result;
}`
}
