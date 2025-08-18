import os
from github import Github
from github.GithubException import UnknownObjectException
from dotenv import load_dotenv

load_dotenv()


def get_github_client():
    token = os.getenv("GITHUB_TOKEN")
    if not token:
        raise RuntimeError("❌ No GITHUB_TOKEN found in environment.")
    return Github(token)


def get_or_create_repo(repo_name: str, private: bool = False):
    """
    Returns existing repo if found, otherwise creates it.
    """
    g = get_github_client()
    user = g.get_user()

    try:
        repo = user.get_repo(repo_name)
        return repo, False  # False = not newly created
    except UnknownObjectException:
        repo = user.create_repo(repo_name, private=private)
        return repo, True  # True = newly created
