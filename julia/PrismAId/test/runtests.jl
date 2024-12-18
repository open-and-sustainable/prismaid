using Test
using PrismAId

@testset "PrismAId Tests" begin
    # Test that run_review raises an error on empty input
    @test_throws ArgumentError PrismAId.run_review("")

    # More tests are difficult because tests are run in Go original code, and are hard to use here too. 
    # But we keep this as placeholder for a forthcoming testing of library call without TOML parsing.
    # @test PrismAId.run_review("Test input") == "Test output"
end
