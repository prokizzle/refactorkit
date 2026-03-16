#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Premium Next.js Design Search - BM25 search engine for web UI/UX style guides.
Forked from ui-ux-pro-max, adapted for shadcn/ui + Tailwind v4 web context.

Usage: python3 design-search.py "query here" [--domain <domain>] [--max-results 3] [--json]
"""

import argparse
import csv
import io
import json
import re
import sys
from collections import defaultdict
from math import log
from pathlib import Path

# Force UTF-8 for stdout/stderr
if sys.stdout.encoding and sys.stdout.encoding.lower() != 'utf-8':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')
if sys.stderr.encoding and sys.stderr.encoding.lower() != 'utf-8':
    sys.stderr = io.TextIOWrapper(sys.stderr.buffer, encoding='utf-8')

# ============ CONFIGURATION ============
DATA_DIR = Path(__file__).resolve().parent.parent / "data"
MAX_RESULTS = 3

CSV_CONFIG = {
    "style": {
        "file": "styles.csv",
        "search_cols": ["Style Category", "Keywords", "Best For", "Type", "AI Prompt Keywords"],
        "output_cols": [
            "Style Category", "Type", "Keywords", "Primary Colors",
            "Effects & Animation", "Best For", "Performance", "Accessibility",
            "Framework Compatibility", "Complexity", "AI Prompt Keywords",
            "CSS/Technical Keywords", "Implementation Checklist",
            "Design System Variables"
        ]
    },
    "color": {
        "file": "colors.csv",
        "search_cols": ["Product Type", "Notes"],
        "output_cols": [
            "Product Type", "Primary", "On Primary", "Secondary", "On Secondary",
            "Accent", "On Accent", "Background", "Foreground", "Card",
            "Card Foreground", "Muted", "Muted Foreground", "Border",
            "Destructive", "On Destructive", "Ring", "Notes"
        ]
    },
    "ux": {
        "file": "ux-guidelines.csv",
        "search_cols": ["Category", "Issue", "Description", "Platform"],
        "output_cols": [
            "Category", "Issue", "Platform", "Description",
            "Do", "Don't", "Code Example Good", "Code Example Bad", "Severity"
        ]
    },
    "typography": {
        "file": "typography.csv",
        "search_cols": ["Font Pairing Name", "Category", "Mood/Style Keywords", "Best For", "Heading Font", "Body Font"],
        "output_cols": [
            "Font Pairing Name", "Category", "Heading Font", "Body Font",
            "Mood/Style Keywords", "Best For", "Google Fonts URL",
            "CSS Import", "Tailwind Config", "Notes"
        ]
    },
    "react": {
        "file": "react-performance.csv",
        "search_cols": ["Category", "Issue", "Keywords", "Description"],
        "output_cols": [
            "Category", "Issue", "Platform", "Description",
            "Do", "Don't", "Code Example Good", "Code Example Bad", "Severity"
        ]
    },
}


# ============ BM25 IMPLEMENTATION ============
class BM25:
    """BM25 ranking algorithm for text search."""

    def __init__(self, k1=1.5, b=0.75):
        self.k1 = k1
        self.b = b
        self.corpus = []
        self.doc_lengths = []
        self.avgdl = 0
        self.idf = {}
        self.doc_freqs = defaultdict(int)
        self.N = 0

    def tokenize(self, text):
        """Lowercase, split, remove punctuation, filter short words."""
        text = re.sub(r'[^\w\s]', ' ', str(text).lower())
        return [w for w in text.split() if len(w) > 2]

    def fit(self, documents):
        """Build BM25 index from documents."""
        self.corpus = [self.tokenize(doc) for doc in documents]
        self.N = len(self.corpus)
        if self.N == 0:
            return
        self.doc_lengths = [len(doc) for doc in self.corpus]
        self.avgdl = sum(self.doc_lengths) / self.N

        for doc in self.corpus:
            seen = set()
            for word in doc:
                if word not in seen:
                    self.doc_freqs[word] += 1
                    seen.add(word)

        for word, freq in self.doc_freqs.items():
            self.idf[word] = log((self.N - freq + 0.5) / (freq + 0.5) + 1)

    def score(self, query):
        """Score all documents against query, return sorted (idx, score) pairs."""
        query_tokens = self.tokenize(query)
        scores = []

        for idx, doc in enumerate(self.corpus):
            s = 0
            doc_len = self.doc_lengths[idx]
            term_freqs = defaultdict(int)
            for word in doc:
                term_freqs[word] += 1

            for token in query_tokens:
                if token in self.idf:
                    tf = term_freqs[token]
                    idf_val = self.idf[token]
                    numerator = tf * (self.k1 + 1)
                    denominator = tf + self.k1 * (1 - self.b + self.b * doc_len / self.avgdl)
                    s += idf_val * numerator / denominator

            scores.append((idx, s))

        return sorted(scores, key=lambda x: x[1], reverse=True)


# ============ SEARCH FUNCTIONS ============
def _load_csv(filepath):
    """Load CSV and return list of dicts."""
    with open(filepath, 'r', encoding='utf-8') as f:
        return list(csv.DictReader(f))


def _search_csv(filepath, search_cols, output_cols, query, max_results):
    """Core search function using BM25."""
    if not filepath.exists():
        return []

    data = _load_csv(filepath)
    documents = [" ".join(str(row.get(col, "")) for col in search_cols) for row in data]

    bm25 = BM25()
    bm25.fit(documents)
    ranked = bm25.score(query)

    results = []
    for idx, score in ranked[:max_results]:
        if score > 0:
            row = data[idx]
            results.append({col: row.get(col, "") for col in output_cols if col in row})

    return results


def detect_domain(query):
    """Auto-detect the most relevant domain from query."""
    query_lower = query.lower()

    domain_keywords = {
        "color": [
            "color", "palette", "hex", "rgb", "token", "semantic", "accent",
            "destructive", "muted", "foreground", "theme"
        ],
        "style": [
            "style", "design", "ui", "minimalism", "glassmorphism", "neumorphism",
            "brutalism", "dark mode", "flat", "aurora", "css", "tailwind",
            "dashboard", "saas", "modern", "landing"
        ],
        "ux": [
            "ux", "usability", "accessibility", "wcag", "touch", "scroll",
            "animation", "keyboard", "navigation", "mobile", "aria", "focus"
        ],
        "typography": [
            "font", "typography", "heading", "body font", "serif", "sans",
            "pairing", "typeface"
        ],
        "react": [
            "react", "next.js", "nextjs", "suspense", "memo", "usecallback",
            "useeffect", "rerender", "bundle", "waterfall", "barrel",
            "dynamic import", "rsc", "server component", "performance",
            "optimization", "hydration", "streaming"
        ],
    }

    scores = {}
    for domain, keywords in domain_keywords.items():
        scores[domain] = sum(
            1 for kw in keywords if re.search(r'\b' + re.escape(kw) + r'\b', query_lower)
        )
    best = max(scores, key=scores.get)
    return best if scores[best] > 0 else "style"


def search(query, domain=None, max_results=MAX_RESULTS):
    """Main search function with auto-domain detection."""
    if domain is None:
        domain = detect_domain(query)

    config = CSV_CONFIG.get(domain, CSV_CONFIG["style"])
    filepath = DATA_DIR / config["file"]

    if not filepath.exists():
        return {"error": f"File not found: {filepath}", "domain": domain}

    results = _search_csv(filepath, config["search_cols"], config["output_cols"], query, max_results)

    return {
        "domain": domain,
        "query": query,
        "file": config["file"],
        "count": len(results),
        "results": results
    }


def format_output(result):
    """Format results for Claude consumption (token-optimized).

    Output suggests shadcn/ui components and Tailwind v4 tokens where applicable.
    """
    if "error" in result:
        return f"Error: {result['error']}"

    output = []
    output.append("## Premium Next.js Design Search Results")
    output.append(f"**Domain:** {result['domain']} | **Query:** {result['query']}")
    output.append(f"**Source:** {result['file']} | **Found:** {result['count']} results")
    output.append("**Stack:** shadcn/ui + Tailwind v4 + Next.js App Router\n")

    for i, row in enumerate(result['results'], 1):
        output.append(f"### Result {i}")
        for key, value in row.items():
            value_str = str(value)
            if len(value_str) > 300:
                value_str = value_str[:300] + "..."
            output.append(f"- **{key}:** {value_str}")
        output.append("")

    # Add shadcn/ui + Tailwind v4 reminder
    if result['domain'] == 'color':
        output.append("> **Tailwind v4 note:** Use CSS custom properties via `@theme` for color tokens.")
        output.append("> Map these to shadcn/ui semantic tokens in `globals.css`.")
    elif result['domain'] == 'style':
        output.append("> **Implementation:** Use shadcn/ui components as the base layer.")
        output.append("> Apply style overrides via Tailwind v4 utility classes and `@theme` tokens.")
    elif result['domain'] == 'typography':
        output.append("> **Tailwind v4 note:** Configure font families via `@theme { --font-heading: ...; }`")
        output.append("> Use `next/font` for automatic font optimization.")

    return "\n".join(output)


# ============ CLI ENTRY POINT ============
if __name__ == "__main__":
    parser = argparse.ArgumentParser(
        description="Premium Next.js Design Search - BM25 engine for web UI/UX"
    )
    parser.add_argument("query", help="Search query")
    parser.add_argument(
        "--domain", "-d",
        choices=list(CSV_CONFIG.keys()),
        help="Search domain (auto-detected if omitted)"
    )
    parser.add_argument(
        "--max-results", "-n",
        type=int, default=MAX_RESULTS,
        help="Max results (default: 3)"
    )
    parser.add_argument(
        "--json", action="store_true",
        help="Output as JSON"
    )

    args = parser.parse_args()

    result = search(args.query, args.domain, args.max_results)

    if args.json:
        print(json.dumps(result, indent=2, ensure_ascii=False))
    else:
        print(format_output(result))
