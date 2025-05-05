import platform
import ctypes
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
    if architecture == "arm64" or architecture == "ARM64":
        lib = CDLL(__file__.replace("__init__.py", "libprismaid_darwin_arm64.dylib"))
    else:
        raise OSError(f"Unsupported architecture for macOS: {architecture}")

else:
    raise OSError(f"Unsupported operating system: {system}")

# Define the low-level function signatures
_RunReviewPython = lib.RunReviewPython
_RunReviewPython.argtypes = [c_char_p]
_RunReviewPython.restype = c_char_p

_DownloadZoteroPDFsPython = lib.DownloadZoteroPDFsPython
_DownloadZoteroPDFsPython.argtypes = [c_char_p, c_char_p, c_char_p, c_char_p]
_DownloadZoteroPDFsPython.restype = c_char_p

_DownloadURLListPython = lib.DownloadURLListPython
_DownloadURLListPython.argtypes = [c_char_p]
_DownloadURLListPython.restype = None  # This function returns void

_ConvertPython = lib.ConvertPython
_ConvertPython.argtypes = [c_char_p, c_char_p]
_ConvertPython.restype = c_char_p

_FreeCString = lib.FreeCString
_FreeCString.argtypes = [c_char_p]
_FreeCString.restype = None

# Python-friendly wrapper functions
def review(toml_configuration):
    """
    Run the PrismAId review process with the given TOML configuration.

    Args:
        toml_configuration (str): TOML configuration as a string

    Raises:
        Exception: If the review process fails
    """
    result = _RunReviewPython(toml_configuration.encode('utf-8'))
    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)

def download_zotero_pdfs(username, api_key, collection_name, parent_dir):
    """
    Download PDFs from Zotero.

    Args:
        username (str): Zotero username
        api_key (str): Zotero API key
        collection_name (str): Name of the Zotero collection
        parent_dir (str): Directory to save the PDFs

    Raises:
        Exception: If the download process fails
    """
    result = _DownloadZoteroPDFsPython(
        username.encode('utf-8'),
        api_key.encode('utf-8'),
        collection_name.encode('utf-8'),
        parent_dir.encode('utf-8')
    )

    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)

def download_url_list(path):
    """
    Download files from URLs listed in a file.

    Args:
        path (str): Path to the file containing URLs
    """
    _DownloadURLListPython(path.encode('utf-8'))

def convert(input_dir, selected_formats):
    """
    Convert files to specified formats.

    Args:
        input_dir (str): Directory containing files to convert
        selected_formats (str): Comma-separated list of target formats

    Raises:
        Exception: If the conversion process fails
    """
    result = _ConvertPython(
        input_dir.encode('utf-8'),
        selected_formats.encode('utf-8')
    )

    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)
