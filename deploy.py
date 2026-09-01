import paramiko
import os
import sys

HOST = "vpn.798698.xyz"
USER = "root"
PASSWORD = "??YMY158879"
LOCAL_BINARY = r"e:\idea\Xboard-Go\Xboard-Go\bin\xboard-go-linux"
REMOTE_PATH = "/opt/xboard-go/xboard-go"
SSH_KEY_FILE = os.path.expanduser("~/.ssh/id_ed25519.pub")

def main():
    ssh = paramiko.SSHClient()
    ssh.set_missing_host_key_policy(paramiko.AutoAddPolicy())

    print(f"[1/5] Connecting to {USER}@{HOST} ...")
    try:
        ssh.connect(HOST, username=USER, password=PASSWORD, timeout=15)
        print("      Connected.")
    except Exception as e:
        print(f"      Connection failed: {e}")
        sys.exit(1)

    # Set up SSH key for future use
    if os.path.exists(SSH_KEY_FILE):
        print("[1.5] Setting up SSH key for future auto-deploy ...")
        with open(SSH_KEY_FILE) as f:
            pubkey = f.read().strip()
        cmd = f'mkdir -p ~/.ssh && grep -qF "{pubkey}" ~/.ssh/authorized_keys 2>/dev/null || echo "{pubkey}" >> ~/.ssh/authorized_keys && chmod 600 ~/.ssh/authorized_keys'
        stdin, stdout, stderr = ssh.exec_command(cmd)
        stdout.read()
        print("      SSH key deployed.")

    # Stop service
    print("[2/5] Stopping xboard-go service ...")
    stdin, stdout, stderr = ssh.exec_command("systemctl stop xboard-go")
    err = stderr.read().decode()
    out = stdout.read().decode()
    if err and "Warning" not in err:
        print(f"      Warning: {err.strip()}")
    print(f"      Service stopped. {out.strip()}")

    # Upload binary via SFTP
    print("[3/5] Uploading binary ...")
    sftp = ssh.open_sftp()
    local_size = os.path.getsize(LOCAL_BINARY)
    print(f"      Local binary size: {local_size:,} bytes")

    # Upload to temp path first, then move
    tmp_path = "/opt/xboard-go/xboard-go.new"
    sftp.put(LOCAL_BINARY, tmp_path)
    remote_size = sftp.stat(tmp_path).st_size
    print(f"      Uploaded {remote_size:,} bytes")

    if remote_size != local_size:
        print("      ERROR: Size mismatch!")
        sftp.close()
        ssh.close()
        sys.exit(1)

    # Move and set permissions
    print("[4/5] Replacing binary and starting service ...")
    stdin, stdout, stderr = ssh.exec_command(
        f"mv {tmp_path} {REMOTE_PATH} && chmod +x {REMOTE_PATH} && systemctl start xboard-go"
    )
    err = stderr.read().decode()
    if err and "Warning" not in err:
        print(f"      Warning: {err.strip()}")

    import time
    time.sleep(2)

    # Check status
    print("[5/5] Checking service status ...")
    stdin, stdout, stderr = ssh.exec_command("systemctl is-active xboard-go")
    status = stdout.read().decode().strip()
    print(f"      Service status: {status}")

    stdin, stdout, stderr = ssh.exec_command("systemctl status xboard-go --no-pager -l 2>&1 | head -15")
    print(stdout.read().decode())

    # Quick health check
    print("[Check] Health endpoint ...")
    stdin, stdout, stderr = ssh.exec_command("curl -s -o /dev/null -w '%{http_code}' http://localhost:8080/healthz")
    http_code = stdout.read().decode().strip()
    print(f"      HTTP health check: {http_code}")

    sftp.close()
    ssh.close()
    print("\n=== Deployment complete ===")

if __name__ == "__main__":
    main()
