import ctypes
import json
import platform
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

_DownloadZoteroPython = lib.DownloadZoteroPython
_DownloadZoteroPython.argtypes = [c_char_p]
_DownloadZoteroPython.restype = c_char_p

_DownloadURLListPython = lib.DownloadURLListPython
_DownloadURLListPython.argtypes = [c_char_p]
_DownloadURLListPython.restype = c_char_p

_ConvertPython = lib.ConvertPython
_ConvertPython.argtypes = [c_char_p, c_char_p, c_char_p, c_char_p, c_char_p]
_ConvertPython.restype = c_char_p

_ScreeningPython = lib.ScreeningPython
_ScreeningPython.argtypes = [c_char_p]
_ScreeningPython.restype = c_char_p

_ValidateConfigPython = lib.ValidateConfigPython
_ValidateConfigPython.argtypes = [c_char_p, c_char_p]
_ValidateConfigPython.restype = c_char_p

_CheckConformancePython = lib.CheckConformancePython
_CheckConformancePython.argtypes = [c_char_p, c_char_p]
_CheckConformancePython.restype = c_char_p

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
    result = cast(bytes | None, _RunReviewPython(toml_configuration.encode("utf-8")))
    if result:
        error_message = ctypes.string_at(result).decode("utf-8")
        _FreeCString(result)
        raise Exception(error_message)


def download_zotero(toml_configuration: str) -> None:
    """
    Download PDFs from Zotero with a TOML configuration.

    Args:
        toml_configuration (str): TOML configuration as a string containing a
            [zotero] table and optionally a [revaise] block.

    Raises:
        Exception: If the download process fails
    """
    result = cast(bytes | None, _DownloadZoteroPython(toml_configuration.encode("utf-8")))

    if result:
        error_message = ctypes.string_at(result).decode("utf-8")
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
    result = cast(bytes | None, _DownloadURLListPython(path.encode("utf-8")))
    if result:
        error_message = ctypes.string_at(result).decode("utf-8")
        _FreeCString(result)
        raise Exception(error_message)


def convert(
    input_dir: str,
    selected_formats: str,
    tika_address: str = "",
    single_file: str = "",
    ocr_only: bool = False,
) -> None:
    """
    Convert files to specified formats.

    Args:
        input_dir (str): Directory containing files to convert
        selected_formats (str): Comma-separated list of target formats
        tika_address (str, optional): Tika server address for OCR fallback (e.g., 'localhost:9998').
                                      Empty string disables OCR fallback. Defaults to "".
        single_file (str, optional): Convert only the specified PDF (PDF format only). Defaults to "".
        ocr_only (bool, optional): Force OCR for PDFs via Tika (PDF format only). Requires tika_address.
                                  Defaults to False.

    Raises:
        Exception: If the conversion process fails
    """
    ocr_only_value = "true" if ocr_only else "false"
    result = cast(
        bytes | None,
        _ConvertPython(
            input_dir.encode("utf-8"),
            selected_formats.encode("utf-8"),
            tika_address.encode("utf-8"),
            single_file.encode("utf-8"),
            ocr_only_value.encode("utf-8"),
        ),
    )

    if result:
        error_message = ctypes.string_at(result).decode("utf-8")
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
    result = cast(bytes | None, _ScreeningPython(toml_configuration.encode("utf-8")))
    if result:
        error_message = ctypes.string_at(result).decode("utf-8")
        _FreeCString(result)
        raise Exception(error_message)


def validate_config(config_type: str, toml_configuration: str) -> None:
    """
    Validate a PrismAId configuration without executing it.

    Args:
        config_type (str): Which configuration schema to validate against.
            One of "review", "screening", or "zotero".
        toml_configuration (str): TOML configuration as a string.

    Raises:
        Exception: If the configuration is invalid or the config_type is
            unrecognized.
    """
    result = cast(
        bytes | None,
        _ValidateConfigPython(
            config_type.encode("utf-8"),
            toml_configuration.encode("utf-8"),
        ),
    )
    if result:
        error_message = ctypes.string_at(result).decode("utf-8")
        _FreeCString(result)
        raise Exception(error_message)


def check_conformance(record_json: str, protocol: str) -> dict:
    """
    Check whether a RevAIse review record conforms to a reporting protocol.

    The verdict and messages come from the protocol's SHACL shapes published by
    the RevAIse model, so conformance is decided by the shapes rather than
    asserted by the tool.

    Args:
        record_json (str): The RevAIse review record as a JSON string.
        protocol (str): The protocol identifier (e.g. "prisma-2020").

    Returns:
        dict: The conformance report with keys "protocol", "conforms", and
            "violations" (each violation carrying a "message").

    Raises:
        Exception: If the check fails, for example an unknown protocol or an
            unreadable record.
    """
    result = cast(
        bytes | None,
        _CheckConformancePython(
            record_json.encode("utf-8"),
            protocol.encode("utf-8"),
        ),
    )
    if not result:
        raise Exception("conformance check returned no result")
    raw = ctypes.string_at(result).decode("utf-8")
    _FreeCString(result)
    report = json.loads(raw)
    if isinstance(report, dict) and report.get("error"):
        raise Exception(report["error"])
    return report
