#!/usr/bin/env python3
"""Login-only curl check derived from Zepp Life 6.12.0's passport SDK."""
import getpass
import json
import subprocess
import uuid
from urllib.parse import parse_qs, urlencode, urlsplit

# Public default network key/IV extracted and verified against the APK native library.
ACCOUNT_KEY = b"xeNtBVqzDc6tuNTh"
ACCOUNT_IV = b"MAAAYAAAAAAAAABg"

REDIRECT = "https://s3-us-west-2.amazonaws.com/hm-registration/successsignin.html"


def account_crypt(data, decrypt=False):
    command = ["openssl", "enc", "-aes-128-cbc", "-K", ACCOUNT_KEY.hex(), "-iv", ACCOUNT_IV.hex()]
    if decrypt:
        command.append("-d")
    result = subprocess.run(command, input=data, capture_output=True)
    if result.returncode:
        raise RuntimeError("Zepp network encryption/decryption failed")
    return result.stdout


def post(url, form, headers=(), encrypted=False):
    # Supply secrets on stdin, not in curl's process arguments. Never follow redirects.
    payload = urlencode(form, doseq=True).encode()
    if encrypted:
        payload = account_crypt(payload)
        headers = [*headers, "X-HM-EKV: 1"]
    result = subprocess.run(
        ["curl", "--silent", "--show-error", "--max-time", "60", "--include",
         "--header", "Content-Type: application/x-www-form-urlencoded",
         *[arg for h in headers for arg in ("--header", h)],
         "--data-binary", "@-", url],
        input=payload, capture_output=True,
    )
    if result.returncode:
        raise RuntimeError("curl request failed (exit %d)" % result.returncode)
    raw = result.stdout
    status, response_headers = None, {}
    while raw.startswith(b"HTTP/"):
        separator = b"\r\n\r\n" if b"\r\n\r\n" in raw else b"\n\n"
        header, found, raw = raw.partition(separator)
        if not found:
            raise RuntimeError("curl returned incomplete HTTP headers")
        lines = header.decode("iso-8859-1").splitlines()
        status = int(lines[0].split()[1])
        response_headers = {k.lower(): v.strip() for k, v in
                            (line.split(":", 1) for line in lines[1:] if ":" in line)}
    if status is None:
        raise RuntimeError("curl returned no HTTP status")
    if encrypted and raw:
        try:
            raw = account_crypt(raw, decrypt=True)
        except RuntimeError:
            # Proxies may return plaintext errors; retain only safe diagnostics.
            pass
    return status, response_headers, raw.decode("utf-8", errors="replace")


def failure_details(status, headers, body):
    """Only expose bounded numeric error codes and response-format metadata."""
    details = {"http_status": status, "authenticated": False,
               "response_bytes": len(body.encode()),
               "authentication_challenge": "www-authenticate" in headers}
    try:
        data = json.loads(body)
    except ValueError:
        details["response_format"] = "html" if body.lstrip().startswith("<") else "text_or_empty"
    else:
        details["response_format"] = "json"
        if isinstance(data, dict):
            for key in ("code", "error_code", "status", "error"):
                value = data.get(key)
                if isinstance(value, int) and not isinstance(value, bool) and abs(value) < 1000000000:
                    details[key] = value
                elif isinstance(value, str) and value.isascii() and value.isdecimal() and len(value) <= 9:
                    details[key] = value
    return details


def main():
    email = input("Zepp account email: ").strip()
    if "@" not in email:
        raise RuntimeError("This check requires a native Zepp email account")
    status, headers, body = post("https://api-user.zepp.com/v2/registrations/tokens", {
        "emailOrPhone": email, "password": getpass.getpass("Zepp password: "),
        "state": "REDIRECTION", "client_id": "HuaMi", "redirect_uri": REDIRECT,
        "region": "us-west-2", "token": ["access", "refresh"], "country_code": "GB",
    }, ["app_name: com.xiaomi.hm.health", "Accept-Language: en-GB",
        "appname: com.xiaomi.hm.health", "cv: 50800_6.12.0", "v: 2.0",
        "appplatform: android_phone",
        "User-Agent: MiFit6.12.0 (Android; Android 6.0; Density/2.0)",
        "X-Request-Id: " + str(uuid.uuid4())], encrypted=True)
    if status not in (302, 303):
        print(json.dumps(failure_details(status, headers, body), indent=2))
        raise RuntimeError("Zepp rejected the email/password request (HTTP %d)" % status)
    location = urlsplit(headers.get("location", ""))
    expected = urlsplit(REDIRECT)
    if (location.scheme, location.netloc, location.path) != (expected.scheme, expected.netloc, expected.path):
        raise RuntimeError("Zepp returned a verification or unexpected redirect")
    query = parse_qs(location.query)
    if query.get("error"):
        raise RuntimeError("Zepp rejected the login; check credentials or verification in the app")
    token = query.get("access", [""])[0]
    if not token:
        raise RuntimeError("Zepp returned no access token")
    status, _, body = post("https://account.zepp.com/v2/client/login", {
        "app_name": "com.xiaomi.hm.health", "app_version": "6.12.0",
        "code": token, "country_code": "GB", "device_id": str(uuid.uuid4()),
        "device_model": "phone", "dn": "api-mifit.zepp.com",
        "grant_type": "access_token", "third_name": "huami", "allow_registration": "false",
    })
    if status != 200:
        raise RuntimeError("Zepp app login failed (HTTP %d)" % status)
    data = json.loads(body)
    info = data.get("token_info") or {}
    print(json.dumps({"result": data.get("result"), "authenticated": bool(
        data.get("result") == "ok" and info.get("app_token") and info.get("user_id"))}, indent=2))


if __name__ == "__main__":
    try:
        main()
    except (RuntimeError, ValueError, OSError) as error:
        print("Login check failed:", error)
        raise SystemExit(1)
