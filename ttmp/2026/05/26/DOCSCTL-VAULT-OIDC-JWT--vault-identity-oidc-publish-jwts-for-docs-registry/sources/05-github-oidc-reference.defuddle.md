---
Title: 05 Github Oidc Reference.Defuddle
DocType: source
Ticket: DOCSCTL-VAULT-OIDC-JWT
Status: active
Topics: [vault, oidc, docsctl, security]
---

## OIDC token claims

To see all the claims supported by GitHub's OIDC provider, review the `claims_supported` entries at [https://token.actions.githubusercontent.com/.well-known/openid-configuration](https://token.actions.githubusercontent.com/.well-known/openid-configuration).

The OIDC token includes the following claims.

### Standard audience, issuer, and subject claims

| Claim | Claim type | Description |
| --- | --- | --- |
| `aud` | Audience | By default, this is the URL of the repository owner, such as the organization that owns the repository. You can set a custom audience with a toolkit command: [`core.getIDToken(audience)`](https://www.npmjs.com/package/@actions/core/v/1.6.0) |
| `iss` | Issuer | The issuer of the OIDC token: `https://token.actions.githubusercontent.com` |
| `sub` | Subject | Defines the subject claim that is to be validated by the cloud provider. This setting is essential for making sure that access tokens are only allocated in a predictable way. |

### Additional standard JOSE header parameters and claims

| Header Parameter | Parameter type | Description |
| --- | --- | --- |
| `alg` | Algorithm | The algorithm used by the OIDC provider. |
| `kid` | Key identifier | Unique key for the OIDC token. |
| `typ` | Type | Describes the type of token. This is a JSON Web Token (JWT). |

| Claim | Claim type | Description |
| --- | --- | --- |
| `exp` | Expires at | Identifies the expiry time of the JWT. |
| `iat` | Issued at | The time when the JWT was issued. |
| `jti` | JWT token identifier | Unique identifier for the OIDC token. |
| `nbf` | Not before | JWT is not valid for use before this time. |

### Custom claims provided by GitHub

| Claim | Description |
| --- | --- |
| `actor` | The personal account that initiated the workflow run. |
| `actor_id` | The ID of personal account that initiated the workflow run. |
| `base_ref` | The target branch of the pull request in a workflow run. |
| `check_run_id` | The check run ID of the current job. |
| `environment` | The name of the environment used by the job. If the `environment` claim is included (also via `include_claim_keys`), an environment is required and must be provided. |
| `event_name` | The name of the event that triggered the workflow run. |
| `head_ref` | The source branch of the pull request in a workflow run. |
| `job_workflow_ref` | For jobs using a reusable workflow, the ref path to the reusable workflow. For more information, see [Using OpenID Connect with reusable workflows](https://docs.github.com/en/actions/deployment/security-hardening-your-deployments/using-openid-connect-with-reusable-workflows). |
| `job_workflow_sha` | For jobs using a reusable workflow, the commit SHA for the reusable workflow file. |
| `ref` | *(Reference)* The git ref that triggered the workflow run. |
| `ref_type` | The type of `ref`, for example: "branch". |
| `repository_visibility` | The visibility of the repository where the workflow is running. Accepts the following values: `internal`, `private`, or `public`. |
| `repository` | The repository from where the workflow is running. |
| `repository_id` | The ID of the repository from where the workflow is running. |
| `repository_owner` | The name of the organization in which the `repository` is stored. |
| `repository_owner_id` | The ID of the organization in which the `repository` is stored. |
| `repo_property_*` | Custom properties defined at the organization or enterprise level that are included as claims in the OIDC token, prefixed with `repo_property_`. For more information, see [Including repository custom properties in OIDC tokens](https://docs.github.com/actions/reference/openid-connect-reference#including-repository-custom-properties-in-oidc-tokens). |
| `run_id` | The ID of the workflow run that triggered the workflow. |
| `run_number` | The number of times this workflow has been run. |
| `run_attempt` | The number of times this workflow run has been retried. |
| `runner_environment` | The type of runner used by the job. Accepts the following values: `github-hosted` or `self-hosted`. |
| `workflow` | The name of the workflow. |
| `workflow_ref` | The ref path to the workflow. For example, `octocat/hello-world/.github/workflows/my-workflow.yml@refs/heads/my_branch`. |
| `workflow_sha` | The commit SHA for the workflow file. |

## OIDC claims used to define trust conditions on cloud roles

Audience and subject claims are typically used in combination while setting conditions on the cloud role/resources to scope its access to the GitHub workflows.

- **Audience:** By default, this value uses the URL of the organization or repository owner. This can be used to set a condition that only the workflows in the specific organization can access the cloud role.
- **Subject:** By default, has a predefined format and is a concatenation of some of the key metadata about the workflow, such as the GitHub organization, repository, branch, or associated [`job`](https://docs.github.com/en/actions/using-workflows/workflow-syntax-for-github-actions#jobsjob_idenvironment) environment. See [Example subject claims](https://docs.github.com/actions/reference/openid-connect-reference#example-subject-claims) to see how the subject claim is assembled from concatenated metadata.

If you need more granular trust conditions, you can customize the subject (`sub`) claim that's included with the JWT. For more information, see [Customizing the token claims](https://docs.github.com/actions/reference/openid-connect-reference#customizing-the-token-claims).

There are also many additional claims supported in the OIDC token that can be used for setting these conditions. In addition, your cloud provider could allow you to assign a role to the access tokens, letting you specify even more granular permissions.

## Example subject claims

The following examples demonstrate how to use "Subject" as a condition, and explain how the "Subject" is assembled from concatenated metadata. The [subject](https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims) uses information from the [`job` context](https://docs.github.com/en/actions/learn-github-actions/contexts#job-context), and instructs your cloud provider that access token requests may only be granted for requests from workflows running in specific branches, environments. The following sections describe some common subjects you can use.

### Filtering for a specific environment

The subject claim includes the environment name when the job references an environment.

You can configure a subject that filters for a specific [environment](https://docs.github.com/en/actions/deployment/targeting-different-environments/managing-environments-for-deployment) name. In this example, the workflow run must have originated from a job that has an environment named `Production`, in a repository named `octo-repo` that is owned by the `octo-org` organization:

- Syntax: `repo:ORG-NAME/REPO-NAME:environment:ENVIRONMENT-NAME`
- Example: `repo:octo-org/octo-repo:environment:Production`

### Filtering for pull\_request events

The subject claim includes the `pull_request` string when the workflow is triggered by a pull request event, but only if the job doesn't reference an environment.

You can configure a subject that filters for the [`pull_request`](https://docs.github.com/en/actions/using-workflows/events-that-trigger-workflows#pull_request) event. In this example, the workflow run must have been triggered by a `pull_request` event in a repository named `octo-repo` that is owned by the `octo-org` organization:

- Syntax: `repo:ORG-NAME/REPO-NAME:pull_request`
- Example: `repo:octo-org/octo-repo:pull_request`

### Filtering for a specific branch

The subject claim includes the branch name of the workflow, but only if the job doesn't reference an environment, and if the workflow is not triggered by a pull request event.

You can configure a subject that filters for a specific branch name. In this example, the workflow run must have originated from a branch named `demo-branch`, in a repository named `octo-repo` that is owned by the `octo-org` organization:

- Syntax: `repo:ORG-NAME/REPO-NAME:ref:refs/heads/BRANCH-NAME`
- Example: `repo:octo-org/octo-repo:ref:refs/heads/demo-branch`

### Filtering for a specific tag

The subject claim includes the tag name of the workflow, but only if the job doesn't reference an environment, and if the workflow is not triggered by a pull request event.

You can create a subject that filters for specific tag. In this example, the workflow run must have originated with a tag named `demo-tag`, in a repository named `octo-repo` that is owned by the `octo-org` organization:

- Syntax: `repo:ORG-NAME/REPO-NAME:ref:refs/tags/TAG-NAME`
- Example: `repo:octo-org/octo-repo:ref:refs/tags/demo-tag`

Any `:` within the metadata values will be replaced with `%3A` in the subject claim.

You can configure a subject that includes metadata containing colons. In this example, the workflow run must have originated from a job that has an environment named `Production:V1`, in a repository named `octo-repo` that is owned by the `octo-org` organization:

- Syntax: `repo:ORG-NAME/REPO-NAME:environment:ENVIRONMENT-NAME`
- Example: `repo:octo-org/octo-repo:environment:Production%3AV1`

## Configuring the subject in your cloud provider

To configure the subject in your cloud provider's trust relationship, you must add the subject string to its trust configuration. The following examples demonstrate how various cloud providers can accept the same `repo:octo-org/octo-repo:ref:refs/heads/demo-branch` subject in different ways:

| Cloud provider | Example |
| --- | --- |
| Amazon Web Services | `"token.actions.githubusercontent.com:sub": "repo:octo-org/octo-repo:ref:refs/heads/demo-branch"` |
| Azure | `repo:octo-org/octo-repo:ref:refs/heads/demo-branch` |
| Google Cloud Platform | `(assertion.sub=='repo:octo-org/octo-repo:ref:refs/heads/demo-branch')` |
| HashiCorp Vault | `bound_subject="repo:octo-org/octo-repo:ref:refs/heads/demo-branch"` |

For more information about configuring specific cloud providers, see the guides listed in [Security hardening your deployments](https://docs.github.com/en/actions/how-tos/security-for-github-actions/security-hardening-your-deployments).

## Customizing the token claims

You can security harden your OIDC configuration by customizing the claims that are included with the JWT. These customizations allow you to define more granular trust conditions on your cloud roles when allowing your workflows to access resources hosted in the cloud:

- You can customize values for `audience` claims. See [Customizing the `audience` value](https://docs.github.com/actions/reference/openid-connect-reference#customizing-the-audience-value).
- You can customize the format of your OIDC configuration by setting conditions on the subject (`sub`) claim that require JWT tokens to originate from a specific repository, reusable workflow, or other source.
- You can define granular OIDC policies by using additional OIDC token claims, such as `repository_id` and `repository_visibility`. See [OpenID Connect](https://docs.github.com/en/actions/concepts/security/openid-connect#understanding-the-oidc-token).
- You can include repository custom properties as claims in OIDC tokens, enabling attribute-based access control policies. See [Including repository custom properties in OIDC tokens](https://docs.github.com/actions/reference/openid-connect-reference#including-repository-custom-properties-in-oidc-tokens).

### Customizing the audience value

When you use custom actions in your workflows, those actions may use the GitHub Actions Toolkit to enable you to supply a custom value for the `audience` claim. Some cloud providers also use this in their official login actions to enforce a default value for the `audience` claim. For example, the [GitHub Action for Azure Login](https://github.com/Azure/login/blob/master/action.yml) provides a default `aud` value of `api://AzureADTokenExchange`, or it allows you to set a custom `aud` value in your workflows. For more information on the GitHub Actions Toolkit, see the [OIDC token](https://github.com/actions/toolkit/tree/main/packages/core#oidc-token) section in the documentation.

If you do not want to use the default `aud` value offered by an action, you can provide a custom value for the `audience` claim. This allows you to set a condition that only workflows in a specific repository or organization can access the cloud role. If the action you are using supports this, you can use the `with` keyword in your workflow to pass a custom `aud` value to the action. For more information, see [Metadata syntax reference](https://docs.github.com/en/actions/creating-actions/metadata-syntax-for-github-actions#inputs).

### Including repository custom properties in OIDC tokens

Organization and enterprise admins can select repository custom properties to include as claims in GitHub Actions OIDC tokens. Once a custom property is added to the OIDC configuration, every repository in the organization or enterprise that has a value set for that property will automatically include it in its OIDC tokens. The property name appears in the token prefixed with `repo_property_`.

This allows you to create attribute-based access control (ABAC) policies in your cloud provider that bind directly to your repository metadata, reducing configuration drift and eliminating the need to manage separate access configuration for each repository.

#### Claim format

Each enabled custom property appears as a separate claim in the OIDC token. The claim name is the property name prefixed with `repo_property_`.

| Custom property name | Claim name in OIDC token |
| --- | --- |
| `business_unit` | `repo_property_business_unit` |
| `workspace_id` | `repo_property_workspace_id` |
| `data_classification` | `repo_property_data_classification` |

#### Supported property types

The following custom property types are supported as OIDC claims. The value representation in the token depends on the property type.

| Property type | Example value in OIDC token | Notes |
| --- | --- | --- |
| String | `"repo_property_team": "platform-eng"` | Value appears as a plain string. |
| Single select | `"repo_property_env_tier": "production"` | The selected option appears as a plain string. |
| Multi select | `"repo_property_regions": "us-east-1,eu-west-1"` | Multiple selected values are joined into a single comma-separated string. |
| True/false | `"repo_property_pci_compliant": "true"` | Boolean values appear as the string `"true"` or `"false"`. |

#### Multi-select value representation

When a repository has a multi-select custom property with multiple values selected, the values are joined into a single comma-separated string in the OIDC token. For example, if a repository has a `regions` property with the values `us-east-1` and `eu-west-1`, the claim appears as:

```json
{
  "repo_property_regions": "us-east-1,eu-west-1"
}
```

When configuring trust policies in your cloud provider, use string matching or contains checks to evaluate multi-select claims.

#### Prerequisites for including custom properties

- Custom properties must already be defined at the organization or enterprise level. For more information, see [Managing custom properties for repositories in your organization](https://docs.github.com/en/organizations/managing-organization-settings/managing-custom-properties-for-repositories-in-your-organization).
- You must be an organization admin or enterprise admin.
- After adding a custom property to the OIDC configuration, all repositories in the organization or enterprise that have a value set for that property will automatically include it in their OIDC tokens.

#### Adding a custom property to OIDC token claims

You can manage which custom properties are included in OIDC tokens using the settings UI or the REST API.

- **Using the settings UI:**
	Navigate to your organization's or enterprise's Actions OIDC settings to view and configure which custom properties are included in OIDC tokens.
- **Using the REST API:**
	To add a custom property to your organization's OIDC token claims, send a `POST` request to the appropriate OIDC custom-property inclusion endpoint. For example:
	- For an organization: `POST /orgs/{org}/actions/oidc/customization/properties/repo`
		- For an enterprise: `POST /enterprises/{enterprise}/actions/oidc/customization/properties/repo` For request parameters and full details, see the REST API documentation for managing OIDC custom properties: [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc).

#### Example token with custom properties

After a custom property is added to the OIDC configuration, repositories with a value set for that property will include it in their tokens. In the following example, two custom properties (`business_unit` and `workspace_id`) are included in the token:

```json
{
  "sub": "repo:my-org/my-repo:ref:refs/heads/main",
  "aud": "https://github.com/my-org",
  "repository": "my-org/my-repo",
  "repo_property_business_unit": "payments",
  "repo_property_workspace_id": "ws-abc123"
}
```

You can use these `repo_property_*` claims as conditions in your cloud provider's trust policy. For an example, see [Example: Filtering on a repository custom property](https://docs.github.com/actions/reference/openid-connect-reference#example-filtering-on-a-repository-custom-property).

### Customizing the subject claims for an organization or repository

To help improve security, compliance, and standardization, you can customize the standard claims to suit your required access conditions. If your cloud provider supports conditions on subject claims, you can create a condition that checks whether the `sub` value matches the path of the reusable workflow, such as `"job_workflow_ref:octo-org/octo-automation/.github/workflows/oidc.yml@refs/heads/main"`. The exact format will vary depending on your cloud provider's OIDC configuration. To configure the matching condition on GitHub, you can use the REST API to require that the `sub` claim must always include a specific custom claim, such as `job_workflow_ref`. You can use the REST API to apply a customization template for the OIDC subject claim; for example, you can require that the `sub` claim within the OIDC token must always include a specific custom claim, such as `job_workflow_ref`. For more information, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc).

Customizing the claims results in a new format for the entire `sub` claim, which replaces the default predefined `sub` format in the token described in [Example subject claims](https://docs.github.com/actions/reference/openid-connect-reference#example-subject-claims).

The following example templates demonstrate various ways to customize the subject claim. To configure these settings on GitHub, admins use the REST API to specify a list of claims that must be included in the subject (`sub`) claim.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

To customize your subject claims, you should first create a matching condition in your cloud provider's OIDC configuration, before customizing the configuration using the REST API. Once the configuration is completed, each time a new job runs, the OIDC token generated during that job will follow the new customization template. If the matching condition doesn't exist in the cloud provider's OIDC configuration before the job runs, the generated token might not be accepted by the cloud provider, since the cloud conditions may not be synchronized.

#### Example: Allowing repository based on visibility and owner

This example template allows the `sub` claim to have a new format, using `repository_owner` and `repository_visibility`:

```json
{
   "include_claim_keys": [
       "repository_owner",
       "repository_visibility"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require that claims must include specific values for `repository_owner` and `repository_visibility`. For example: `"sub": "repository_owner:monalisa:repository_visibility:private"`. The approach lets you restrict cloud role access to only private repositories within an organization or enterprise.

#### Example: Allowing access to all repositories with a specific owner

This example template enables the `sub` claim to have a new format with only the value of `repository_owner`.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
   "include_claim_keys": [
       "repository_owner"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require that claims must include a specific value for `repository_owner`. For example: `"sub": "repository_owner:monalisa"`

#### Example: Requiring a reusable workflow

This example template allows the `sub` claim to have a new format that contains the value of the `job_workflow_ref` claim. This enables an enterprise to use reusable workflows to enforce consistent deployments across its organizations and repositories.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
     "include_claim_keys": [
         "job_workflow_ref"
     ]
  }
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require that claims must include a specific value for `job_workflow_ref`. For example: `"sub": "job_workflow_ref:octo-org/octo-automation/.github/workflows/oidc.yml@refs/heads/main"`.

#### Example: Requiring a reusable workflow and other claims

The following example template combines the requirement of a specific reusable workflow with additional claims.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

This example also demonstrates how to use `"context"` to define your conditions. This is the part that follows the repository in the default `sub` format. For example, when the job references an environment, the context contains: `environment:ENVIRONMENT-NAME`.

```json
{
   "include_claim_keys": [
       "repo",
       "context",
       "job_workflow_ref"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require that claims must include specific values for `repo`, `context`, and `job_workflow_ref`.

This customization template requires that the `sub` uses the following format: `repo:ORG-NAME/REPO-NAME:environment:ENVIRONMENT-NAME:job_workflow_ref:REUSABLE-WORKFLOW-PATH`. For example: `"sub": "repo:octo-org/octo-repo:environment:prod:job_workflow_ref:octo-org/octo-automation/.github/workflows/oidc.yml@refs/heads/main"`

#### Example: Granting access to a specific repository

This example template lets you grant cloud access to all the workflows in a specific repository, across all branches/tags and environments.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
   "include_claim_keys": [
       "repo"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require a `repo` claim that matches the required value.

#### Example: Using system-generated GUIDs

This example template enables predictable OIDC claims with system-generated GUIDs that do not change between renames of entities (such as renaming a repository).

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
     "include_claim_keys": [
         "repository_id"
     ]
  }
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require a `repository_id` claim that matches the required value.

or:

```json
{
   "include_claim_keys": [
       "repository_owner_id"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require a `repository_owner_id` claim that matches the required value.

#### Example: Context value with:

This example demonstrates how to handle context value with `:`. For example, when the job references an environment named `production:eastus`.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
   "include_claim_keys": [
       "environment",
       "repository_owner"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require that claims must include a specific value for `environment` and `repository_owner`. For example: `"sub": "environment:production%3Aeastus:repository_owner:octo-org"`.

#### Example: Filtering on a repository custom property

This example template allows the `sub` claim to include a repository custom property claim. Custom properties included in OIDC tokens appear prefixed with `repo_property_` in the token, but the `include_claim_keys` value uses the full claim name as it appears in the token.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
   "include_claim_keys": [
       "repo_property_workspace_id"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require that claims must include a specific value for `repo_property_workspace_id`. For example: `"sub": "repo_property_workspace_id:ws-abc123"`.

#### Resetting organization template customizations

This example template resets the subject claims to the default format. This template effectively opts out of any organization-level customization policy.

To apply this configuration, submit a request to the API endpoint and include the required configuration in the request body. For organizations, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-an-organization), and for repositories, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
   "include_claim_keys": [
       "repo",
       "context"
   ]
}
```

In your cloud provider's OIDC configuration, configure the `sub` condition to require that claims must include specific values for `repo` and `context`.

#### Resetting repository template customizations

All repositories in an organization have the ability to opt in or opt out of (organization and repository-level) customized `sub` claim templates.

To opt out a repository and reset back to the default `sub` claim format, a repository administrator must use the REST API endpoint at [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

To configure repositories to use the default `sub` claim format, use the `PUT /repos/{owner}/{repo}/actions/oidc/customization/sub` REST API endpoint at with the following request body.

```json
{
   "use_default": true
}
```

#### Example: Configuring a repository to use an organization template

Once an organization has created a customized `sub` claim template, the REST API can be used to programmatically apply the template to repositories within the organization. A repository administrator can configure their repository to use the template created by the administrator of their organization.

To configure the repository to use the organization's template, a repository admin must use the `PUT /repos/{owner}/{repo}/actions/oidc/customization/sub` REST API endpoint at with the following request body. For more information, see [REST API endpoints for GitHub Actions OIDC](https://docs.github.com/en/rest/actions/oidc#set-the-customization-template-for-an-oidc-subject-claim-for-a-repository).

```json
{
   "use_default": false
}
```

## Debugging your OIDC claims

You can use the [`github/actions-oidc-debugger`](https://github.com/github/actions-oidc-debugger) action to visualize the claims that would be sent, before integrating with a cloud provider. This action requests a JWT and prints the claims included within the JWT that were received from GitHub Actions.

## Workflow permissions for the requesting the OIDC token

### Required permission

- The job or workflow must grant the [`id-token: write`](https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#permissions) permission to allow GitHub's OIDC provider to create a JSON Web Token (JWT):
	```yaml
	permissions:
	  id-token: write
	```
- Without `id-token: write`, the OIDC JWT ID token cannot be requested. This setting only enables fetching and setting the OIDC token; it does not grant write access to other resources.

### Setting permissions

- To fetch an OIDC token for a workflow, set the permission at the workflow level:
	```yaml
	permissions:
	  id-token: write # This is required for requesting the JWT
	  contents: read # This is required for actions/checkout
	```
- To fetch an OIDC token for a single job, set the permission within that job:
	```yaml
	permissions:
	  id-token: write # This is required for requesting the JWT
	```
- Additional permissions may be required depending on workflow needs.

### Reusable workflows

- For reusable workflows owned by the same user, organization, or enterprise as the caller, the OIDC token generated in the reusable workflow is accessible from the caller's context.
- For reusable workflows outside your enterprise or organization, set the `permissions` setting for `id-token` to `write` explicitly at the caller workflow or job level. This ensures the OIDC token is only available to intended caller workflows.

## Methods for requesting the OIDC token

Custom actions can request the OIDC token using:

- The `getIDToken()` method from the Actions toolkit. For more information, see [OIDC Token](https://www.npmjs.com/package/@actions/core/v/1.6.0#oidc-token) in the npm package documentation.
- The following environment variables on the runner.
	| Variable | Description |
	| --- | --- |
	| `ACTIONS_ID_TOKEN_REQUEST_URL` | The URL for GitHub's OIDC provider. |
	| `ACTIONS_ID_TOKEN_REQUEST_TOKEN` | Bearer token for the request to the OIDC provider. |
	For example:
	```shell
	curl -H "Authorization: bearer $ACTIONS_ID_TOKEN_REQUEST_TOKEN" "$ACTIONS_ID_TOKEN_REQUEST_URL&audience=api://AzureADTokenExchange"
	```