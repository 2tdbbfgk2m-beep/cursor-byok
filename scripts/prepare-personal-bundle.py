"""Set personal bundle metadata while preserving the build configuration schema."""
import os
from pathlib import Path
import plistlib
import re

version = os.environ['VERSION']
p = Path('bin/Cursor助手.app/Contents/Info.plist')
info = plistlib.loads(p.read_bytes())
info['CFBundleShortVersionString'] = version
info['CFBundleVersion'] = os.environ['GITHUB_RUN_NUMBER']
info['LSUIElement'] = False
p.write_bytes(plistlib.dumps(info))

p = Path('build/config.yml')
text, count = re.subn(r'^([ \t]+version:[ \t]*).*$', lambda m: m[1] + '"' + version + '"', p.read_text(), count=1, flags=re.M)
if count != 1:
    raise ValueError('Expected exactly one application version field')
p.write_text(text)
