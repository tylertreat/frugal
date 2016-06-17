from setuptools import setup, find_packages

from frugal.version import __version__

setup(
    name='frugal',
    version=__version__,
    description='Frugal Python Library',
    maintainer='Messaging Team',
    maintainer_email='messaging@workiva.com',
    url='http://github.com/Workiva/frugal',
    packages=find_packages(exclude=('frugal.tests', 'frugal.tests.*')),
    install_requires=[
        'thrift==0.9.3'
    ],
    extras_require={
        'tornado': ["nats-client==0.2.4"]
    }
)
