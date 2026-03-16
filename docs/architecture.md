# HDC Architecture Diagram

```mermaid
flowchart TD
    subgraph Input
        HF["helmfile.yaml / helmfile.yaml.gotmpl\nhelmfile.d/*.yaml"]
        CFG["Config File / CLI Flags\nEnv Variables"]
    end

    subgraph cmd["cmd (CLI Entry Point)"]
        CLI["hdc check"]
    end

    subgraph internal["internal/"]
        CONFIG["config\n─────────────\nLoad & validate config\nInit logger"]

        PARSER["parser\n─────────────\nStrip Go templates\nParse YAML\nExtract releases & repos"]

        REPOSITORY["repository\n─────────────\nFetch index.yaml\nParse chart metadata\nHTTP client"]

        CHECKER["checker\n─────────────\nSemver comparison\nMaintenance age check\nApply exclusion rules\nGenerate findings"]

        REPORT["report\n─────────────\nJSON / Markdown / HTML\nExit code for CI/CD"]

        MODELS["models\n─────────────\nHelmfile · Release\nRepository · Finding\nResult"]
    end

    subgraph External
        HELM_REPO["Helm Repositories\n(index.yaml)"]
    end

    subgraph Output
        OUT["Report\n(stdout / file)"]
        EXIT["Exit Code\n0 = ok · 1 = issues found"]
    end

    CFG --> CLI
    HF  --> CLI

    CLI --> CONFIG
    CLI --> PARSER
    CLI --> CHECKER
    CLI --> REPORT

    PARSER     -->|"[]Release, []Repository"| CHECKER
    CHECKER    -->|FetchIndex| REPOSITORY
    REPOSITORY -->|HTTP GET| HELM_REPO
    HELM_REPO  -->|index.yaml| REPOSITORY
    REPOSITORY -->|ChartMetadata| CHECKER
    CHECKER    -->|"[]Finding"| REPORT

    REPORT --> OUT
    REPORT --> EXIT

    MODELS -.->|shared types| PARSER
    MODELS -.->|shared types| CHECKER
    MODELS -.->|shared types| REPOSITORY
    MODELS -.->|shared types| REPORT

    CONFIG -.->|Config struct| PARSER
    CONFIG -.->|Config struct| CHECKER
    CONFIG -.->|Config struct| REPORT
```

## Data Flow

1. **CLI** reads config and resolves helmfile path(s)
2. **parser** strips Go template expressions, unmarshals YAML, returns `[]Release` and `[]Repository`
3. **checker** iterates releases, calls **repository** to fetch each repo's `index.yaml`, compares versions and last-updated timestamps against configured thresholds
4. **report** formats `[]Finding` and writes output; exits non-zero if issues are found

## Module Dependency Rules

- `models` has no imports from other internal packages
- `parser`, `repository`, `checker`, `report` import `models` only
- `cmd` is the only package that imports all internal packages
- No circular dependencies permitted
