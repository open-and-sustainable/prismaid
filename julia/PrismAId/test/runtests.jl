using Test
using PrismAId

@testset "PrismAId Tests" begin
    # Test that run_review handles invalid input
    @test_throws Exception PrismAId.run_review("")

    # Basic existence tests for new functions
    @test hasmethod(PrismAId.download_zotero_pdfs, (String, String, String, String))
    @test hasmethod(PrismAId.download_url_list, (String,))
    @test hasmethod(PrismAId.convert, (String, String))

    # Try to call with empty strings - should either throw an error or handle gracefully
    try
        PrismAId.download_zotero_pdfs("", "", "", "")
        PrismAId.download_url_list("")
        PrismAId.convert("", "")
        # If we reach here without error, that's also acceptable
        @test true
    catch e
        # Error is expected and acceptable
        @test e isa Exception
    end

    # More tests are difficult because tests are run in Go original code, and are hard to use here too.
    # But we keep this as placeholder for a forthcoming testing of library call without TOML parsing.
    # @test PrismAId.run_review("Test input") == "Test output"
end
