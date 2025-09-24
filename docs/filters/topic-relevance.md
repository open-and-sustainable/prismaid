# Topic Relevance Filter

## Overview

The topic relevance filter screens manuscripts based on their relevance to user-specified research topics. It calculates a relevance score from 0.0 to 1.0 for each manuscript by analyzing its content against topic descriptions. Manuscripts scoring above the minimum threshold are included; others are filtered out as off-topic.

## How It Works

### Core Concept

When you specify topics like "machine learning in healthcare applications", the filter analyzes each manuscript through three components to determine relevance:

### The Three Component Scores

#### 1. Keyword Match Score (default weight: 40%)
Extracts individual meaningful words from your topic descriptions and searches for them in the manuscript.

**Example:**
- Topic: "machine learning in healthcare applications"
- Extracted keywords: ["machine", "learning", "healthcare", "applications"]
- Searches for these exact words in title, abstract, and keywords
- Excludes stop words like "in", "the", "and"
- Score = (matched keywords / total keywords) × 1.5 (capped at 1.0)

#### 2. Concept Match Score (default weight: 40%)
Identifies multi-word phrases and concepts, capturing domain-specific terminology.

**Example:**
- Topic: "deep learning models for clinical decision support"
- Extracted concepts: ["deep learning", "learning models", "clinical decision", "decision support"]
- Searches for complete phrases in the manuscript
- Catches technical terms that have specific meaning when words appear together

#### 3. Field Relevance Score (default weight: 20%)
Examines journal name, research field, and subject categories to determine domain alignment.

**Example:**
- Papers from "Journal of Medical Artificial Intelligence" get high scores for AI healthcare topics
- Papers from "Agricultural Science Quarterly" get low scores
- Helps filter papers that coincidentally use similar words but are from unrelated fields

### Score Calculation

```
overall_score = (keyword_score × weight1) + (concept_score × weight2) + (field_score × weight3)
```

**Example Scoring:**

Manuscript: "A deep learning approach for medical diagnosis using neural networks"
- Keyword matches: "deep", "learning", "medical", "diagnosis" → Score: 0.8
- Concept matches: "deep learning", "medical diagnosis" → Score: 0.7
- Field: Published in "IEEE Transactions on Medical Imaging" → Score: 0.9
- **Overall score** = (0.8 × 0.4) + (0.7 × 0.4) + (0.9 × 0.2) = 0.78

With `min_score = 0.5`, this manuscript would be included.

## Configuration

### Basic Configuration

```toml
[filters.topic_relevance]
enabled = true
use_ai = false                    # Use rule-based scoring
min_score = 0.5                   # Minimum score threshold (0.0-1.0)

# Define your research topics
topics = [
    "machine learning applications in healthcare",
    "artificial intelligence for medical diagnosis",
    "deep learning models for clinical decision support",
    "natural language processing for electronic health records"
]

# Scoring component weights
[filters.topic_relevance.score_weights]
keyword_match = 0.4               # Weight for keyword matching
concept_match = 0.4               # Weight for concept matching
field_relevance = 0.2             # Weight for field/journal relevance
```

### Configuration Parameters

| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `enabled` | boolean | false | Enable/disable the filter |
| `use_ai` | boolean | false | Use AI for semantic understanding (requires LLM config) |
| `topics` | array | [] | List of topic descriptions |
| `min_score` | float | 0.5 | Minimum relevance score (0.0-1.0) |
| `score_weights.keyword_match` | float | 0.4 | Weight for keyword matching |
| `score_weights.concept_match` | float | 0.4 | Weight for concept matching |
| `score_weights.field_relevance` | float | 0.2 | Weight for field relevance |

## Weight Adjustment Strategies

### Technical/Specific Search
Increase concept_match weight for technical phrase emphasis:
```toml
[filters.topic_relevance.score_weights]
keyword_match = 0.3
concept_match = 0.5    # Higher weight for technical phrases
field_relevance = 0.2
```

### Broad Topic Search
Increase keyword_match weight for individual term focus:
```toml
[filters.topic_relevance.score_weights]
keyword_match = 0.5    # Focus on individual terms
concept_match = 0.3
field_relevance = 0.2
```

### Domain-Specific Search
Increase field_relevance weight for journal/field emphasis:
```toml
[filters.topic_relevance.score_weights]
keyword_match = 0.3
concept_match = 0.3
field_relevance = 0.4  # Emphasize relevant journals
```

## Writing Effective Topics

### Good Topics - Detailed and Specific
```toml
topics = [
    "machine learning algorithms for predicting patient readmission rates",
    "deep learning models for medical image segmentation in radiology",
    "natural language processing for extracting clinical information from doctor notes"
]
```

### Poor Topics - Too Generic
```toml
topics = [
    "AI",
    "machine learning",
    "healthcare"
]
```

## Output Fields

The filter adds these fields to each manuscript record:

| Field | Type | Description |
|-------|------|-------------|
| `topic_relevance_score` | float | Overall relevance score (0.0-1.0) |
| `topic_relevance_confidence` | float | Confidence in the assessment |
| `matched_keywords` | array | List of keywords that matched |
| `matched_concepts` | array | List of concepts that matched |
| `exclusion_reason` | string | Explanation if manuscript is excluded |
| `component_scores` | object | Individual scores for keyword, concept, and field relevance (AI mode) |
| `reasoning` | string | AI's explanation of relevance decision (AI mode only) |

## Score Interpretation

| Score Range | Interpretation | Typical Action |
|------------|----------------|----------------|
| 0.0 - 0.3 | Low relevance | Usually exclude |
| 0.3 - 0.5 | Moderate relevance | Review threshold setting |
| 0.5 - 0.7 | Good relevance | Usually include |
| 0.7 - 1.0 | High relevance | Definitely include |

## Example Use Cases

### Focused Literature Review
```toml
[filters.topic_relevance]
enabled = true
use_ai = false
topics = [
    "machine learning for depression detection from social media",
    "natural language processing for suicide risk assessment",
    "AI-powered chatbots for mental health support",
    "predictive models for psychiatric treatment outcomes"
]
min_score = 0.6  # Require strong relevance
```

### Broad Technology Survey
```toml
[filters.topic_relevance]
enabled = true
use_ai = false
topics = [
    "artificial intelligence in medical imaging and diagnostics",
    "machine learning for drug discovery and development",
    "AI-assisted surgical planning and navigation",
    "predictive analytics for population health management",
    "deep learning for genomics and precision medicine"
]
min_score = 0.4  # Lower threshold for broader coverage
```

### Methodology-Focused Search
```toml
[filters.topic_relevance]
enabled = true
use_ai = false
topics = [
    "transformer models for clinical text analysis",
    "graph neural networks for drug-drug interaction prediction",
    "reinforcement learning for treatment recommendation systems",
    "federated learning for privacy-preserving medical AI"
]
min_score = 0.5

[filters.topic_relevance.score_weights]
keyword_match = 0.5    # Higher weight on technical terms
concept_match = 0.3
field_relevance = 0.2
```

## AI-Enhanced Mode (Optional)

When `use_ai = true`, the filter uses Large Language Models for deep semantic understanding of topic relevance, going beyond keyword matching to understand conceptual relationships and research context.

### AI Configuration

```toml
[filters.topic_relevance]
enabled = true
use_ai = true
min_score = 0.6  # Often higher threshold with AI

# LLM configuration required
[[filters.llm]]
provider = "OpenAI"
api_key = ""  # Uses environment variable if empty
model = "gpt-4o-mini"
temperature = 0.01
```

### AI Capabilities

The AI-powered mode provides:

- **Semantic Understanding**: Recognizes conceptual relationships beyond literal word matches
- **Context Awareness**: Understands research methodology alignment with topics
- **Synonym Recognition**: Identifies related terms and domain-specific vocabulary
- **Interdisciplinary Connections**: Finds relevance across different fields
- **Research Question Alignment**: Evaluates if research objectives match topic interests

### AI Prompt Details

The AI evaluates manuscripts using comprehensive analysis:

**Evaluation Criteria:**
1. Direct keyword matches with the topics
2. Conceptual alignment with the research areas
3. Field/domain relevance
4. Methodological relevance
5. Research questions and objectives alignment

**Structured Response:**
The AI returns detailed scoring with explanations:
```json
{
  "overall_score": 0.75,      // Relevance score from 0.0 to 1.0
  "component_scores": {
    "keyword_match": 0.8,      // Direct keyword alignment
    "concept_match": 0.7,      // Conceptual relationship strength
    "field_relevance": 0.75    // Domain/field alignment
  },
  "matched_keywords": ["machine learning", "healthcare"],
  "matched_concepts": ["predictive modeling", "clinical decision support"],
  "confidence": 0.85,          // AI's confidence in assessment
  "is_relevant": true,         // Boolean relevance decision
  "reasoning": "Strong alignment with AI healthcare topics through predictive modeling approach"
}
```

### Graceful Fallback

The system automatically falls back to rule-based scoring when:
- No LLM models are configured
- API calls fail or timeout
- Response parsing errors occur
- Invalid JSON response from AI

This ensures the filter always produces results, never blocking the screening pipeline.

### AI Mode Examples

#### Interdisciplinary Research Detection
AI mode excels at identifying relevant interdisciplinary work:

```toml
topics = [
    "machine learning for climate change mitigation",
    "AI-driven renewable energy optimization"
]
```

A paper on "Neural Network Control Systems for Wind Turbine Efficiency" would be recognized as relevant through conceptual understanding, even without exact keyword matches.

#### Methodology-Based Relevance
AI understands methodological alignment:

```toml
topics = [
    "deep learning approaches for medical imaging",
    "convolutional neural networks in radiology"
]
```

Papers using CNNs for any medical imaging task would score high, regardless of specific medical terminology used.

### Benefits of AI Mode

1. **Reduced False Negatives**: Catches relevant papers that keyword matching might miss
2. **Contextual Understanding**: Evaluates research questions and objectives, not just terms
3. **Flexible Topic Interpretation**: Natural language topic descriptions work effectively
4. **Confidence Scoring**: Provides reliability metric for each assessment
5. **Detailed Reasoning**: Explains why manuscripts are considered relevant or not

## Performance Considerations

### Rule-Based Mode
- **Speed**: Very fast, processes hundreds of manuscripts per second
- **Accuracy**: Good for clear topic matches with direct keyword alignment
- **Cost**: Free, no API calls required
- **Best for**: Well-defined topics with clear terminology

### AI-Enhanced Mode
- **Speed**: Limited by API rate limits (typically 10-60 requests per minute)
- **Accuracy**: Superior semantic understanding and context awareness
- **Cost**: API costs apply (approximately $0.001-0.005 per manuscript)
- **Best for**: 
  - Interdisciplinary research topics
  - Conceptual or theoretical topics
  - Emerging research areas with evolving terminology
  - High-precision screening requirements

### Optimization Tips

1. **Use AI Selectively**: Apply AI mode after initial filtering to reduce costs
2. **Batch Processing**: The filter supports efficient batch processing
3. **Rate Limit Configuration**: Adjust TPM/RPM limits based on your API tier
4. **Hybrid Approach**: Use rule-based for initial screening, AI for borderline cases

## Filter Order

Topic relevance is applied after:
1. Deduplication
2. Language detection
3. Article type classification

This ensures efficient processing by removing duplicates and non-target language papers first.