# taco

````markdown
# Go Project

## Prerequisites
- Go 1.22+  
- Git  

## Setup: GitHub Token (if needed for private modules or operations requiring GitHub auth)

If your project depends on private repositories, or you have GitHub operations that require authentication, you’ll need a **GitHub Personal Access Token (PAT)**.

1. Go to **GitHub → Settings → Developer settings → Personal access tokens**. :contentReference[oaicite:0]{index=0}  
2. Generate a new token. Choose scopes minimally needed (e.g. `repo` to access private repos). :contentReference[oaicite:1]{index=1}  
3. Copy the token (you’ll only see it once).  
4. On your local machine set it as an environment variable, e.g.:

   ```bash
   export GITHUB_TOKEN=your_token_here
````

5. Use that token for private repo fetches. For example:

   ```bash
   git clone https://username:${GITHUB_TOKEN}@github.com/username/private-repo.git
   ```

For full GitHub docs on creating and managing tokens see: *Managing your personal access tokens* ([GitHub Docs][1])

---

## Install dependencies

```bash
cd <root of project>
go mod tidy
```

---

## Build

Run this from the **root directory** of your project:

```bash
go build -o taco ./cmd/taco
```

This will produce an executable named `taco` in the root (or whatever directory you're in when you run build).

---

## Run

After building:

```bash
./taco init
```

[1]: https://docs.github.com/en/authentication/keeping-your-account-and-data-secure/managing-your-personal-access-tokens?utm_source=chatgpt.com "Managing your personal access tokens"


I wanna test out the CI/CD