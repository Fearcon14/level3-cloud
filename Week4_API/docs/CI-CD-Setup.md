# Week4_API CI/CD on git.onstackit.cloud

When you push changes under `Week4_API/`, the GitLab pipeline on **git.onstackit.cloud** will:

1. **Build** a Linux/amd64 Docker image for the PaaS API
2. **Push** it to `registry.onstackit.cloud/kevin-sinn/paas-api:latest`
3. **Rollout** the deployment on your SKE cluster so the new image is used:
   `kubectl rollout restart deployment/paas-api`

---

## 1. Create CI/CD variables (secrets)

In your project on **git.onstackit.cloud**:

1. Go to **Settings → CI/CD**.
2. Expand **Variables**.
3. Add the following variables. Mark sensitive ones as **Masked** and optionally **Protected** (so they are only available on protected branches).

### Registry (to push the image)

| Key | Type | Value | Masked | Protected |
|-----|------|--------|--------|-----------|
| `STACKIT_REGISTRY_USER` | Variable | Username for `registry.onstackit.cloud` | ✅ | Optional |
| `STACKIT_REGISTRY_PASSWORD` | Variable | Password or token for the registry | ✅ | Optional |

**Where to get these:**
Use the same credentials you use locally for `docker login registry.onstackit.cloud`.
Typically these come from the STACKIT portal (Container Registry / project settings). If you use a service account or token for the registry, use that as the password.

### Cluster access (to run `kubectl rollout restart`)

| Key | Type | Value | Masked | Protected |
|-----|------|--------|--------|-----------|
| `KUBECONFIG` | **File** | Full contents of your kubeconfig file | — | Optional |

**How to set a File variable:**

- Click **Add variable**.
- Key: `KUBECONFIG`.
- Type: **File**.
- Value: paste the **entire** contents of your kubeconfig (e.g. the file you got from Terraform: `kubeconfig-<cluster-name>` or from the STACKIT/SKE console).
- GitLab will store it and in each job expose a path to a temporary file; `kubectl` will read it via the `KUBECONFIG` environment variable.

**Optional:** If `paas-api` is not in the `default` namespace, add:

| Key | Type | Value |
|-----|------|--------|
| `K8S_NAMESPACE` | Variable | Your namespace (e.g. `paas`) |

---

## 2. Summary of credentials you need

- **STACKIT_REGISTRY_USER** – Registry username (e.g. from STACKIT Container Registry / project).
- **STACKIT_REGISTRY_PASSWORD** – Registry password or token.
- **KUBECONFIG** – Full kubeconfig file content (File-type variable) so the runner can run `kubectl` against your SKE cluster.

---

## 3. When the pipeline runs

The pipeline is defined in the .forgejo/workflows/ directory: **`api-image.yml`**.

- It runs only when there are **changes under `Week4_API/`** (on push or on merge request).
- The runner must have **network access** to:
  - `registry.onstackit.cloud` (push image),
  - Your **Kubernetes API server** (for `kubectl rollout restart`).
  If the cluster API is not public, the GitLab runner must run in a network that can reach it (e.g. same VPC or a runner inside the cluster).

---

## 4. Quick check

After adding the variables:

1. Push a small change under `Week4_API/` (e.g. a comment in a Go file or in the Dockerfile).
2. Open **CI/CD → Pipelines** and confirm the pipeline runs.
3. Check that the **build-and-push** job succeeds and the **deploy-rollout** job runs `kubectl rollout restart deployment/paas-api` and then `kubectl rollout status ...` successfully.
