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

function download_zotero_pdfs(username::String, api_key::String, collection_name::String, parent_dir::String)
    # Validate inputs
    if isempty(username) || isempty(api_key) || isempty(collection_name) || isempty(parent_dir)
        throw(ArgumentError("All parameters must be non-empty strings"))
    end

    # Call the C function
    c_output = ccall((:DownloadZoteroPDFsPython, library_path), Ptr{Cchar},
                    (Cstring, Cstring, Cstring, Cstring),
                    username, api_key, collection_name, parent_dir)

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

# Export public functions
export run_review, download_zotero_pdfs, download_url_list, convert, screening

end # module PrismAId
