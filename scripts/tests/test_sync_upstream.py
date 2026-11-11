import os
from pathlib import Path
import subprocess
import tempfile
import unittest

SCRIPT = Path(__file__).resolve().parents[1] / 'sync-upstream.sh'

class UpstreamSyncTests(unittest.TestCase):
    def setUp(self):
        self.temp = tempfile.TemporaryDirectory()
        self.addCleanup(self.temp.cleanup)
        root = Path(self.temp.name)
        self.upstream, self.origin, self.checkout = root/'upstream', root/'origin.git', root/'personal'
        self.git(root, 'init', '-b', 'main', str(self.upstream))
        self.configure(self.upstream)
        (self.upstream/'model.txt').write_text('base\n')
        self.commit(self.upstream, 'base')
        self.git(root, 'clone', '--bare', str(self.upstream), str(self.origin))
        self.git(root, 'clone', str(self.origin), str(self.checkout))
        self.configure(self.checkout)
        self.git(self.checkout, 'remote', 'add', 'upstream', str(self.upstream))
        (self.checkout/'provider.txt').write_text('personal provider names\n')
        self.commit(self.checkout, 'personal feature')
        self.git(self.checkout, 'push', 'origin', 'main')
        self.old_origin = self.git(self.origin, 'rev-parse', 'main')

    def git(self, cwd, *args):
        return subprocess.check_output(['git',*args],cwd=cwd,stderr=subprocess.DEVNULL,text=True).strip()

    def configure(self,cwd):
        for key,value in [('user.name','Sync Test'),('user.email','test@example.invalid'),('commit.gpgsign','false')]:
            self.git(cwd,'config',key,value)

    def commit(self,cwd,message):
        self.git(cwd,'add','.')
        self.git(cwd,'commit','-m',message)

    def sync(self):
        output = Path(self.temp.name)/'output'
        result = subprocess.run(['bash',str(SCRIPT)],cwd=self.checkout,env={**os.environ,'GITHUB_OUTPUT':str(output)},capture_output=True,text=True)
        # The merge script must never push before the workflow's regression checks.
        self.assertEqual(self.git(self.origin,'rev-parse','main'),self.old_origin)
        return result, output.read_text() if output.exists() else ''

    def test_unchanged_upstream_is_noop(self):
        head = self.git(self.checkout,'rev-parse','HEAD')
        result,output = self.sync()
        self.assertEqual(result.returncode,0,result.stderr)
        self.assertIn('changed=false',output)
        self.assertEqual(self.git(self.checkout,'rev-parse','HEAD'),head)

    def test_merge_preserves_personal_feature_without_pushing(self):
        (self.upstream/'upstream.txt').write_text('upstream fix\n')
        self.commit(self.upstream,'upstream fix')
        result,output = self.sync()
        self.assertEqual(result.returncode,0,result.stderr)
        self.assertIn('changed=true',output)
        self.assertEqual((self.checkout/'provider.txt').read_text(),'personal provider names\n')
        self.assertEqual((self.checkout/'upstream.txt').read_text(),'upstream fix\n')

    def test_conflict_aborts_without_changing_published_branch(self):
        (self.checkout/'model.txt').write_text('personal model routing\n')
        self.commit(self.checkout,'personal routing')
        head = self.git(self.checkout,'rev-parse','HEAD')
        (self.upstream/'model.txt').write_text('upstream conflicting routing\n')
        self.commit(self.upstream,'upstream routing')
        result,output = self.sync()
        self.assertEqual(result.returncode,1,result.stderr)
        self.assertNotIn('changed=true',output)
        self.assertEqual(self.git(self.checkout,'rev-parse','HEAD'),head)
        self.assertFalse((self.checkout/'.git/MERGE_HEAD').exists())
        self.assertEqual(self.git(self.checkout,'status','--porcelain'),'')

if __name__ == '__main__':
    unittest.main()
