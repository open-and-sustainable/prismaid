module PRISMAID

using Libdl

function get_library_path()
    lib_dir = joinpath(@__DIR__, "..", "deps", "lib")
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
const lib = Libdl.dlopen(library_path)

# Use `lib` with `ccall` as needed
# Example:
# function my_function(arg1::Int)
#     ccall((:function_in_lib, lib), ReturnType, (Cint,), arg1)
# end

end # module PRISMAID
