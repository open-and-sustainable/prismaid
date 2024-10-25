import platform
from ctypes import CDLL, c_char_p

# Determine the system and load the correct shared library
system = platform.system()
if system == "Linux":
    lib = CDLL(__file__.replace("__init__.py", "libprismaid_linux_amd64.so"))
elif system == "Windows":
    lib = CDLL(__file__.replace("__init__.py", "libprismaid_windows_amd64.dll"))
elif system == "Darwin":
    lib = CDLL(__file__.replace("__init__.py", "libprismaid_darwin_amd64.dylib"))
else:
    raise OSError("Unsupported operating system")

# Example function from the shared library
RunReviewPython = lib.RunReviewPython
RunReviewPython.argtypes = [c_char_p]
RunReviewPython.restype = c_char_p
