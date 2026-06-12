import sys
import subprocess
import shutil

def check_command(cmd, name, optional=False, required_for_demo=False):
    path = shutil.which(cmd)
    if path:
        try:
            version = subprocess.check_output([cmd, "--version"], stderr=subprocess.STDOUT).decode("utf-8").split("\n")[0]
            print(f"[{'x'}] {name} found: {version}")
        except Exception:
            print(f"[{'x'}] {name} found at {path}")
    else:
        if required_for_demo:
            print(f"[ ] {name} missing (Required for preferred demo path)")
        elif optional:
            print(f"[ ] {name} missing (Optional)")
        else:
            print(f"[ ] {name} missing (Useful for native development)")

def main():
    print("Checking local tooling...")
    print("--- Required for preferred demo path ---")
    check_command("task", "Task", required_for_demo=True)
    check_command("az", "Azure CLI", required_for_demo=True)
    
    print("\n--- Optional / Native tools ---")
    check_command("python3", "Python")
    check_command("git", "Git")
    check_command("go", "Go")
    check_command("node", "Node")
    check_command("npm", "npm")
    check_command("docker", "Docker", optional=True)

if __name__ == "__main__":
    main()
