import os
import yaml

from lang.base import LanguageBase


class Go(LanguageBase):
    """
    Go implementation of LanguageBase.
    """

    def update_frugal(self, version, root):
        """
        Update the go version. Go versioning is controlled by git tags, so
        there is nothing to do here.
        """
        os.chdir('{0}/examples/go'.format(root))

        with open('glide.yaml') as f:
            glide_yaml = yaml.load(f.read())
            for imp in glide_yaml['import']:
                if imp['package'] == 'github.com/Workiva/frugal':
                    imp['version'] = version

        with open('glide.yaml', 'w') as f:
            yaml.dump(glide_yaml, f, default_flow_style=False)

    def update_expected_tests(self, root):
        pass
