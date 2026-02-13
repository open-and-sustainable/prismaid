using Test
using PrismAId

function is_unsafe_convert_cstring_error(e)
    return e isa MethodError && occursin("unsafe_convert", sprint(showerror, e)) && occursin("Cstring", sprint(showerror, e))
end

function assert_no_cstring_wrapper_error(f)
    try
        f()
        @test true
    catch e
        @test !is_unsafe_convert_cstring_error(e)
        @test e isa Exception
    end
end

@testset "PrismAId Tests" begin
    # Test that run_review handles invalid input
    @test_throws Exception PrismAId.run_review("")

    # Basic existence tests for new functions
    @test hasmethod(PrismAId.download_zotero_pdfs, (String, String, String, String))
    @test hasmethod(PrismAId.download_url_list, (String,))
    @test hasmethod(PrismAId.convert, (String, String))
    @test hasmethod(PrismAId.screening, (String,))

    # Regression: wrapper calls should not fail on pointer conversion/free path
    # with MethodError(unsafe_convert(::Type{Ptr{Int8}}, ::Cstring)).
    minimal_screening_toml = """
    [project]
    name = "test"
    input_file = "missing.csv"
    output_file = "out.csv"
    text_column = "abstract"
    identifier_column = "id"
    output_format = "csv"

    [filters.language]
    enabled = false
    accepted_languages = ["en"]

    [filters.article_type]
    enabled = false
    accepted_types = ["research_article"]

    [filters.deduplication]
    enabled = false
    use_ai = false
    compare_fields = ["title"]

    [filters.topic_relevance]
    enabled = false
    use_ai = false
    topics = []
    """

    # run_review may fail for invalid project config, but should not fail on wrapper type mismatch.
    assert_no_cstring_wrapper_error(() -> PrismAId.run_review("[project]\nname=\"x\"\n"))

    # download wrappers may fail with argument/runtime errors, but not wrapper type mismatch.
    assert_no_cstring_wrapper_error(() -> PrismAId.download_zotero_pdfs("u", "k", "g", "/nonexistent"))
    assert_no_cstring_wrapper_error(() -> PrismAId.download_url_list("/nonexistent/urls.txt"))

    # convert wrapper may fail for missing path/Tika, but not wrapper type mismatch.
    assert_no_cstring_wrapper_error(() -> PrismAId.convert("/nonexistent", "pdf"))

    # screening wrapper may fail for runtime/config/file reasons, but not wrapper type mismatch.
    assert_no_cstring_wrapper_error(() -> PrismAId.screening(minimal_screening_toml))

    # More tests are difficult because tests are run in Go original code, and are hard to use here too.
    # But we keep this as placeholder for a forthcoming testing of library call without TOML parsing.
    # @test PrismAId.run_review("Test input") == "Test output"
end
