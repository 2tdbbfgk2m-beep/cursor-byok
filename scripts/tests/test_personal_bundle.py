import os
from pathlib import Path
import plistlib
import subprocess
import sys
import tempfile
import unittest

REPO = Path(__file__).resolve().parents[2]

class PersonalBundleTests(unittest.TestCase):
    def test_release_version_updates_info_without_changing_schema(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            config = root/'build/config.yml'
            config.parent.mkdir()
            # The actual Wails config has both a schema version and info.version.
            config.write_text('version: "3"\ninfo:\n  companyName: "Cursor助手"\n  version: "1.0.1"\n')
            info_path = root/'bin/Cursor助手.app/Contents/Info.plist'
            info_path.parent.mkdir(parents=True)
            info_path.write_bytes(plistlib.dumps({'CFBundleIdentifier':'com.cursor.wuxianxubei','CFBundleShortVersionString':'0.0.50'}))
            result = subprocess.run([sys.executable,str(REPO/'scripts/prepare-personal-bundle.py')],cwd=root,env={**os.environ,'VERSION':'1.0.1-personal.6','GITHUB_RUN_NUMBER':'6'},capture_output=True,text=True)
            self.assertEqual(result.returncode,0,result.stderr)
            version = subprocess.check_output(['go','run','./scripts/release','version','-config',str(config)],cwd=REPO,text=True)
            self.assertEqual(version,'1.0.1-personal.6')
            self.assertEqual(config.read_text().splitlines()[0],'version: "3"')
            self.assertIn('companyName: "Cursor助手"',config.read_text())
            info = plistlib.loads(info_path.read_bytes())
            self.assertEqual(info['CFBundleIdentifier'],'com.cursor.wuxianxubei')
            self.assertEqual(info['CFBundleShortVersionString'],'1.0.1-personal.6')
            self.assertEqual(info['CFBundleVersion'],'6')

if __name__ == '__main__':
    unittest.main()
