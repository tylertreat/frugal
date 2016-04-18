class LanguageBase(object):
    """Language update implementations must implement LanguageBase."""

    def update_frugal(self, version, root, dry_run=False):
        """Update the frugal version."""
        raise Exception('update_frugal not implemented')
