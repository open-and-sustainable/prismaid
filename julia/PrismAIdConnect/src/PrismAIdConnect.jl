module PrismAIdConnect

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
    c_output = ccall((:RunReviewPython, library_path), Cstring, (Cstring,), input)
    if c_output == C_NULL
        throw(RuntimeError("The C shared library returned a null pointer."))
    end

    result = unsafe_string(c_output)
    
    # Free the C string if necessary
    ccall((:FreeCString, library_path), Cvoid, (Ptr{Cchar},), c_output)
    
    return result
end

end # module PrismAIdConnect
