# `hatch`

`hatch` uploads (currently only pull request-related) webhooks to cloud storage to a bucket specified in the config file as
`hatch.gcs_bucket`.

`hatch` uploads to a path of the form `<prefix>/<repo-fullname>/<pull-request-number>/YYYY/M[M]/D[D]/<hook GUID>`,
where `<prefix>` can be specified in the config file as `hatch.gcs_prefix`. Because of the way GitHub reports a repository
fullname, this will be part of the GCS object tree nicely as `organization/repository` or `owner/repository`.

It should not be depended upon that `hatch` exclusively stores pull-request related hooks, as it may be extended in future
to store other types of webhook.

`hatch` is configured by passing in flags, which include `port`, `config-path`, `webhook-hmac-secret-path`, `gcs-auth-path`.
