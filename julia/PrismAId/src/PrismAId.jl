module PrismAId

function get_library_path()
    lib_dir = joinpath(@__DIR__, "..", "deps")
    arch = Sys.ARCH
    if Sys.islinux()
        return joinpath(lib_dir, "linux-$arch", "libprismaid_linux_amd64.so")
    elseif Sys.iswindows()
        return joinpath(lib_dir, "windows-$arch", "libprismaid_windows_amd64.dll")
    elseif Sys.isapple()
        return joinpath(lib_dir, "macos-$arch", "libprismaid_macos_arm64.dylib")
    else
        error("Unsupported platform or architecture")
    end
end

const library_path = get_library_path()

function run_review(input::String)
    # Validate input
    if isempty(input)
        throw(ArgumentError("Input cannot be empty"))
    end

    # Call the C function, passing the String directly
    c_output = ccall((:RunReviewPython, library_path), Ptr{Cchar}, (Cstring,), input)
    if c_output == C_NULL
        return nothing
    end

    result = unsafe_string(c_output)

    # Free the C string if necessary
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)

    return result
end

"""
    download_zotero(input::String)

Download PDF attachments from a Zotero collection using a TOML configuration
string. The configuration must include a `[zotero]` table with `user`,
`api_key`, `group`, and `output_dir`; it may also include an optional
`[revaise]` block to update a RevAIse review record. Returns `nothing` on
success and throws an exception when the shared library reports an error.
"""
function download_zotero(input::String)
    # Validate input
    if isempty(input)
        throw(ArgumentError("Input cannot be empty"))
    end

    # Call the C function
    c_output = ccall((:DownloadZoteroPython, library_path), Ptr{Cchar}, (Cstring,), input)

    if c_output == C_NULL
        return nothing  # Success case returns NULL/nil in Python interface
    end

    # If we got here, it's an error message
    result = unsafe_string(c_output)

    # Free the C string
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)

    throw(ErrorException(result))
end

function download_url_list(path::String)
    # Validate input
    if isempty(path)
        throw(ArgumentError("Path cannot be empty"))
    end

    # Call the C function
    c_output = ccall((:DownloadURLListPython, library_path), Ptr{Cchar}, (Cstring,), path)

    if c_output == C_NULL
        return nothing  # Success case returns NULL/nil in Python interface
    end

    # If we got here, it's an error message
    result = unsafe_string(c_output)

    # Free the C string
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)

    throw(ErrorException(result))
end

function convert(input_dir::String, selected_formats::String, tika_address::String="", single_file::String="", ocr_only::Bool=false)
    # Validate inputs
    if isempty(input_dir)
        throw(ArgumentError("Input directory cannot be empty"))
    end

    if isempty(selected_formats)
        throw(ArgumentError("Selected formats cannot be empty"))
    end

    # tika_address can be empty string to disable OCR fallback
    ocr_only_value = ocr_only ? "true" : "false"

    # Call the C function
    c_output = ccall((:ConvertPython, library_path), Ptr{Cchar},
                    (Cstring, Cstring, Cstring, Cstring, Cstring),
                    input_dir, selected_formats, tika_address, single_file, ocr_only_value)

    if c_output == C_NULL
        return nothing  # Success case returns NULL/nil in Python interface
    end

    # If we got here, it's an error message
    result = unsafe_string(c_output)

    # Free the C string
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)

    throw(ErrorException(result))
end

function screening(input::String)
    # Validate input
    if isempty(input)
        throw(ArgumentError("Input cannot be empty"))
    end

    # Call the C function
    c_output = ccall((:ScreeningPython, library_path), Ptr{Cchar}, (Cstring,), input)

    if c_output == C_NULL
        return nothing  # Success case returns NULL/nil in Python interface
    end

    # If we got here, it's an error message
    result = unsafe_string(c_output)

    # Free the C string
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)

    throw(ErrorException(result))
end

"""
    validate_config(config_type::String, input::String)

Validate a PrismAId TOML configuration without executing it. `config_type`
selects the configuration schema and must be `"review"`, `"screening"`, or
`"zotero"`. `input` is the TOML configuration string. Returns `nothing` when the
configuration is valid and throws an exception describing the problem otherwise.
"""
function validate_config(config_type::String, input::String)
    # Validate inputs
    if isempty(config_type)
        throw(ArgumentError("Config type cannot be empty"))
    end

    # Call the C function
    c_output = ccall((:ValidateConfigPython, library_path), Ptr{Cchar}, (Cstring, Cstring), config_type, input)

    if c_output == C_NULL
        return nothing  # Success case returns NULL/nil in Python interface
    end

    # If we got here, it's an error message
    result = unsafe_string(c_output)

    # Free the C string
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)

    throw(ErrorException(result))
end

"""
    check_conformance(record_json::String, protocol::String)

Check whether a RevAIse review record conforms to a reporting protocol. `record_json`
is the RevAIse review record as a JSON string and `protocol` selects the protocol
(for example `"prisma-2020"`). The verdict and messages come from the protocol's
SHACL shapes published by the RevAIse model.

Returns the conformance report as a JSON string — an object with `protocol`,
`conforms`, and `violations` (each carrying a `message`), or an `error` field when
the check fails. Parse it with a JSON package such as JSON.jl.
"""
function check_conformance(record_json::String, protocol::String)
    if isempty(protocol)
        throw(ArgumentError("Protocol cannot be empty"))
    end

    c_output = ccall((:CheckConformancePython, library_path), Ptr{Cchar}, (Cstring, Cstring), record_json, protocol)
    if c_output == C_NULL
        throw(ErrorException("conformance check returned no result"))
    end

    result = unsafe_string(c_output)
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)
    return result
end

"""
    protocol_guidance(protocol::String)

Return a protocol's full requirement checklist. `protocol` selects the protocol
(for example `"prisma-2020"`). The checklist is extracted from the protocol's SHACL
shapes published by the RevAIse model; it is advisory and does not constrain the
order in which prismAId's tools are used.

Returns the guidance as a JSON string — an object with `protocol`, `name`,
`version`, `status`, and `requirements` (each with a `target_class` and a
`message`), or an `error` field when guidance fails. Parse it with a JSON package
such as JSON.jl.
"""
function protocol_guidance(protocol::String)
    if isempty(protocol)
        throw(ArgumentError("Protocol cannot be empty"))
    end

    c_output = ccall((:ProtocolGuidancePython, library_path), Ptr{Cchar}, (Cstring,), protocol)
    if c_output == C_NULL
        throw(ErrorException("protocol guidance returned no result"))
    end

    result = unsafe_string(c_output)
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)
    return result
end

"""
    generate_revaise_record(params_json::String)

Build a seed RevAIse review record from a JSON parameters object. `params_json`
is a JSON string such as
`{"title":"...","authors":["..."],"type":"SYSTEMATIC_REVIEW","status":"PROTOCOL","include_manual_stage_stubs":true}`.

Returns the seed review record as a JSON string, suitable for writing to the
configuration's record_file, or an object with an `error` field when generation
fails.
"""
function generate_revaise_record(params_json::String)
    if isempty(params_json)
        throw(ArgumentError("Parameters cannot be empty"))
    end

    c_output = ccall((:GenerateRevAIseRecordPython, library_path), Ptr{Cchar}, (Cstring,), params_json)
    if c_output == C_NULL
        throw(ErrorException("record generation returned no result"))
    end

    result = unsafe_string(c_output)
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)
    return result
end

"""
    revaise_schema(params_json::String)

Serve the RevAIse data model from the released, verified artifacts. `params_json`
is a JSON string such as `{"type":"SearchStage"}` to describe a type, `{}` to list
the available classes and enums, `{"raw":true}` for the full JSON Schema, or
`{"context":true}` for the JSON-LD context. Artifacts are fetched live; the LinkML
source is never used.

Returns the result as a JSON string, or an object with an `error` field on failure.
"""
function revaise_schema(params_json::String)
    if isempty(params_json)
        throw(ArgumentError("Parameters cannot be empty"))
    end

    c_output = ccall((:RevAIseSchemaPython, library_path), Ptr{Cchar}, (Cstring,), params_json)
    if c_output == C_NULL
        throw(ErrorException("revaise schema returned no result"))
    end

    result = unsafe_string(c_output)
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)
    return result
end

"""
    merge_record_stage(record_json::String, stage_json::String)

Merge a stage into an existing RevAIse review record. `stage_json` is a JSON
object with at least a `stage_type`; it fills a matching stub (matched by
stage_type and stage_label) or is appended when none matches.

Returns the updated review record as a JSON string, or an object with an `error`
field on failure.
"""
function merge_record_stage(record_json::String, stage_json::String)
    c_output = ccall((:MergeRecordStagePython, library_path), Ptr{Cchar}, (Cstring, Cstring), record_json, stage_json)
    if c_output == C_NULL
        throw(ErrorException("record merge returned no result"))
    end

    result = unsafe_string(c_output)
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)
    return result
end

"""
    validate_record(record_json::String)

Validate a RevAIse review record against the data-model JSON Schema, fetched live.
This checks structural validity (field names, types, required slots), distinct
from `check_conformance` which checks a reporting protocol.

Returns the result as a JSON string — an object with `valid` and `errors` — or an
object with an `error` field when the schema cannot be retrieved.
"""
function validate_record(record_json::String)
    c_output = ccall((:ValidateRecordPython, library_path), Ptr{Cchar}, (Cstring,), record_json)
    if c_output == C_NULL
        throw(ErrorException("record validation returned no result"))
    end

    result = unsafe_string(c_output)
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)
    return result
end

# Export public functions
export run_review, download_zotero, download_url_list, convert, screening, validate_config, check_conformance, protocol_guidance, generate_revaise_record, revaise_schema, merge_record_stage, validate_record

end # module PrismAId
