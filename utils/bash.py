import os
import subprocess
import shutil


def run_bash(script, cwd=None):
    bash_path = shutil.which("bash") or "C:\\Program Files\\Git\\bin\\bash.exe"
    if cwd:
        script_path = os.path.join(os.getcwd(), script)  # absolute path
    else:
        script_path = script
    if not os.path.exists(script_path):
        raise FileNotFoundError(f"❌ Script not found: {script}")
    print(f"⚡ Running bash script: {script}")
    subprocess.run([bash_path, script_path], check=True, cwd=cwd)
