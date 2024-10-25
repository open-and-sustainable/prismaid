from setuptools import setup, find_packages
import platform
import subprocess
import os

# Function to determine the version
def get_version():
    try:
        # Get the latest tag
        version = subprocess.check_output(["git", "describe", "--tags", "--abbrev=0"]).decode().strip()
    except subprocess.CalledProcessError:
        # If no tags are found, fallback to the short commit hash
        version = subprocess.check_output(["git", "rev-parse", "--short", "HEAD"]).decode().strip()
    return version

# Set the package version
package_version = get_version()

# Determine the correct shared library based on the OS
system = platform.system()
shared_libs = [
    "libprismaid_linux_amd64.so",
    "libprismaid_windows_amd64.dll",
    "libprismaid_darwin_amd64.dylib",
]

# Use the README from the root directory
with open("../README.md", "r") as fh:
    long_description = fh.read()

setup(
    name="prismaid",
    version=package_version,
    description="PrismAId library package",
    long_description=long_description,
    long_description_content_type="text/markdown",
    packages=find_packages(where="package"),
    package_dir={"": "package"},
    package_data={"": shared_libs},
    include_package_data=True,
    classifiers=[
        "Programming Language :: Python :: 3",
        "Operating System :: OS Independent",
    ],
    install_requires=[],
)
