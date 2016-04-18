import os

from yaml import dump, load

from lang.base import LanguageBase


class Dart(LanguageBase):
    """
    Dart implementation of LanguageBase. Uses PyYAML to parse all
    pubspec.yaml's.
    """

    def update_frugal(self, version, root, dry_run=False):
        """Update the dart version."""
        # Update libary pubspec
        def update_lib(data):
            data['version'] = version
        os.chdir('{0}/lib/dart'.format(root))
        self._update(update_lib, 'Dart lib', dry_run)

        # Update example pubspec
        def update_example(data):
            data['dependencies']['frugal']['version'] = '^{0}'.format(
                version
            )
        os.chdir('{0}/example/dart/browser'.format(root))
        self._update(update_example, 'Dart example', dry_run)

    def _update(self, update, prefix, dry_run):
        """
        Update pubspec.yaml in current directory using the given update
        function.
        """
        pubspec = 'pubspec.yaml'
        with open(pubspec, 'r') as f:
            data = load(f.read())
            update(data)
        if dry_run:
            print '{0} pubspec.yaml'.format(prefix)
            print dump(data, default_flow_style=False)
            return

        with open(pubspec, 'w') as f:
            dump(data, f, default_flow_style=False)

