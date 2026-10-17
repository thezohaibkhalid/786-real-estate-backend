import importlib.util
import os
import tempfile
import unittest
from pathlib import Path

spec = importlib.util.spec_from_file_location('configure', Path(__file__).with_name('configure.py'))
configure = importlib.util.module_from_spec(spec)
spec.loader.exec_module(configure)


class ConfigurationTests(unittest.TestCase):
    def test_one_env_updates_all_apps_and_preserves_literal_credentials(self):
        with tempfile.TemporaryDirectory() as tmp:
            root = Path(tmp)
            (root / 'deploy').mkdir()
            (root / 'deploy/nginx.conf').write_text(Path(__file__).with_name('nginx.conf').read_text())
            text = ('PUBLIC_SITE_URL=https://site.example.test\nADMIN_SITE_URL=https://admin.example.test\n'
                    'DATABASE_URL=postgres://user:literal$#password@db.example.test/app\n'
                    'SESSION_SECRET=' + 'a' * 64 + '\nREVALIDATE_SECRET=' + 'b' * 64 + '\n'
                    'EMAIL_SERVER_PASSWORD=secret$#with spaces\nR2_SECRET_ACCESS_KEY=r2-secret\n')
            (root / '.env').write_text(text)
            nginx = root / 'nginx.conf'
            configure.render(root, nginx, root / 'missing-cert')
            backend = configure.load_env(root / 'env/backend.env')
            frontend = configure.load_env(root / 'env/frontend.env')
            admin = configure.load_env(root / 'env/admin.env')
            self.assertEqual(backend['EMAIL_SERVER_PASSWORD'], 'secret$#with spaces')
            self.assertEqual(backend['ADMIN_ORIGIN'], 'https://admin.example.test')
            self.assertEqual(frontend['SITE_URL'], admin['PUBLIC_SITE_URL'])
            self.assertEqual(frontend['REVALIDATE_SECRET'], admin['REVALIDATE_SECRET'])
            self.assertNotIn('DATABASE_URL', frontend)
            self.assertNotIn('R2_SECRET_ACCESS_KEY', admin)
            self.assertEqual(os.stat(root / 'env/backend.env').st_mode & 0o777, 0o600)
            (root / '.env').write_text(text.replace('site.example.test', 'new.example.test'))
            configure.render(root, nginx, root / 'missing-cert')
            self.assertIn('server_name new.example.test;', nginx.read_text())
            self.assertEqual(configure.load_env(root / 'env/frontend.env')['SITE_URL'], 'https://new.example.test')
            self.assertEqual(configure.load_env(root / 'env/admin.env')['PUBLIC_SITE_URL'], 'https://new.example.test')

    def test_production_origin_rejects_paths_and_injection(self):
        for value in ['http://site.test', 'https://site.test/path', 'https://user@site.test', 'https://site.test:8443', 'https://site.test;return 200']:
            with self.subTest(value=value):
                with self.assertRaises(ValueError):
                    configure.origin(value, 'SITE_URL')


if __name__ == '__main__':
    unittest.main()
