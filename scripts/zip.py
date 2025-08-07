import os
import stat
import sys
import time
import zipfile

ZIPNAME = sys.argv[1]
FILENAME = sys.argv[2]

with open(FILENAME, "rb") as f:
    data = f.read()

with zipfile.ZipFile(ZIPNAME, "w") as zfile:
    info = zipfile.ZipInfo(FILENAME)
    info.date_time = time.localtime(time.time())[:6]
    info.compress_type = zipfile.ZIP_DEFLATED
    info.compress_level = 9
    info.create_system = 3  # unix
    info.external_attr = (stat.S_IFREG | 0o755) << 16
    zfile.writestr(info, data)

os.remove(FILENAME)
