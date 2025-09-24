import platform
import ctypes
from ctypes import CDLL, c_char_p
from typing import cast

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
_DownloadURLListPython.restype = c_char_p

_ConvertPython = lib.ConvertPython
_ConvertPython.argtypes = [c_char_p, c_char_p]
_ConvertPython.restype = c_char_p

_ScreeningPython = lib.ScreeningPython
_ScreeningPython.argtypes = [c_char_p]
_ScreeningPython.restype = c_char_p

_FreeCString = lib.FreeCString
_FreeCString.argtypes = [c_char_p]
_FreeCString.restype = None

# Python-friendly wrapper functions
def review(toml_configuration: str) -> None:
    """
    Run the PrismAId review process with the given TOML configuration.

    Args:
        toml_configuration (str): TOML configuration as a string

    Raises:
        Exception: If the review process fails
    """
    result = cast(bytes | None, _RunReviewPython(toml_configuration.encode('utf-8')))
    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)

def download_zotero_pdfs(username: str, api_key: str, collection_name: str, parent_dir: str) -> None:
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
    result = cast(bytes | None, _DownloadZoteroPDFsPython(
        username.encode('utf-8'),
        api_key.encode('utf-8'),
        collection_name.encode('utf-8'),
        parent_dir.encode('utf-8')
    ))

    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)

def download_url_list(path: str) -> None:
    """
    Download files from URLs listed in a file.

    Args:
        path (str): Path to the file containing URLs

    Raises:
        Exception: If the file cannot be opened or read
    """
    result = cast(bytes | None, _DownloadURLListPython(path.encode('utf-8')))
    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)

def convert(input_dir: str, selected_formats: str) -> None:
    """
    Convert files to specified formats.

    Args:
        input_dir (str): Directory containing files to convert
        selected_formats (str): Comma-separated list of target formats

    Raises:
        Exception: If the conversion process fails
    """
    result = cast(bytes | None, _ConvertPython(
        input_dir.encode('utf-8'),
        selected_formats.encode('utf-8')
    ))

    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)

def screening(toml_configuration: str) -> None:
    """
    Run the PrismAId screening process to filter manuscripts based on various criteria.

    Args:
        toml_configuration (str): TOML configuration as a string containing:
            - Project settings (name, input/output files, etc.)
            - Filter configurations (deduplication, language, article type, topic relevance)
            - Optional LLM settings for AI-assisted screening

    Raises:
        Exception: If the screening process fails
    """
    result = cast(bytes | None, _ScreeningPython(toml_configuration.encode('utf-8')))
    if result:
        error_message = ctypes.string_at(result).decode('utf-8')
        _FreeCString(result)
        raise Exception(error_message)
