import platform
from ctypes import CDLL, c_char_p

# Determine the system and load the correct shared library
system = platform.system()
architecture = platform.machine().lower()

# Load the correct shared library based on system and architecture
if system == "Linux":
    if architecture == "amd64" or architecture == "x86_64":
        lib = CDLL(__file__.replace("__init__.py", "libprismaid_linux_amd64.so"))
    else:
        raise OSError(f"Unsupported architecture for Linux: {architecture}")

elif system == "Windows":
    if architecture == "amd64" or architecture == "x86_64":
        lib = CDLL(__file__.replace("__init__.py", "libprismaid_windows_amd64.dll"))
    else:
        raise OSError(f"Unsupported architecture for Windows: {architecture}")

elif system == "Darwin":
    if architecture == "amd64" or architecture == "x86_64":
        lib = CDLL(__file__.replace("__init__.py", "libprismaid_darwin_amd64.dylib"))
    else:
        raise OSError(f"Unsupported architecture for macOS: {architecture}")

else:
    raise OSError(f"Unsupported operating system: {system}")

# Example function from the shared library
RunReviewPython = lib.RunReviewPython
RunReviewPython.argtypes = [c_char_p]
RunReviewPython.restype = c_char_p
