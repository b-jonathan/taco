from pathlib import Path
import subprocess
import typer
from utils.github import get_or_create_repo
import InquirerPy.inquirer as iq
from utils.bash import run_bash

app = typer.Typer()


def github_init():
    """
    CLI command to create a GitHub repo if it doesn't already exist.
    """

    repo_name = iq.text(message="Enter the repository name:").execute()

    private = (
        iq.select(
            message="Choose repository visibility:", choices=["Public", "Private"]
        ).execute()
        == "Private"
    )

    repo, created = get_or_create_repo(repo_name, private)
    if created:
        typer.echo(f"🚀 Created repo '{repo_name}': {repo.clone_url}")
    else:
        typer.echo(f"✅ Repo '{repo_name}' already exists: {repo.clone_url}")

    # Ensure local clone
    parent_dir = Path.cwd().parent
    repo_dir = parent_dir / repo_name
    if not repo_dir.exists():
        # Prefer SSH if available; fall back to HTTPS or gh
        remote = getattr(repo, "ssh_url", None) or repo.clone_url
        # If you use GitHub CLI auth, this also works:
        # owner_repo = getattr(repo, "full_name", None)  # e.g. "you/repo"
        # run_bash(f"gh repo clone {owner_repo} {repo_name}")
        subprocess.run(["git", "clone", remote, str(repo_dir)], check=True)

    return repo_dir


def express_init(repo_name):
    """
    CLI command to initialize an Express.js project.
    """
    # Run the bash script to initialize Express.js
    script_path = Path(__file__).parent / "scripts" / "init_express.sh"
    if not script_path.exists():
        raise FileNotFoundError(f"Missing script: {script_path}")

    # Run the initializer script INSIDE the newly cloned repo
    run_bash(script_path, cwd=repo_name)
    typer.echo("✅ Express.js project initialized successfully.")


@app.command()
def init():
    """
    CLI command to initialize an Express.js project.
    """

    repo_dir = github_init()
    stack = iq.select(
        message="Which stack do you want to scaffold?",
        choices=["Express", "Next.js (TODO)"],
    ).execute()

    if stack == "Express":
        express_init(repo_dir)


if __name__ == "__main__":
    app()
