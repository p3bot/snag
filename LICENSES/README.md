# Third-Party Licenses

`snag` is licensed under the Mozilla Public License 2.0 (see ../LICENSE).

This directory contains the licenses for third-party dependencies used in snag.

## Dependencies

### go-rod/rod (MIT License)

- **Purpose**: Chrome DevTools Protocol library for browser automation
- **Repository**: https://github.com/go-rod/rod
- **License**: MIT License
- **License File**: rod.LICENSE

### spf13/cobra (Apache License 2.0)

- **Purpose**: CLI framework for building command-line applications
- **Repository**: https://github.com/spf13/cobra
- **License**: Apache License 2.0
- **License File**: cobra.LICENSE

### JohannesKaufmann/html-to-markdown (MIT License)

- **Purpose**: HTML to Markdown conversion library
- **Repository**: https://github.com/JohannesKaufmann/html-to-markdown
- **License**: MIT License
- **License File**: html-to-markdown.LICENSE

### k3a/html2text (MIT License)

- **Purpose**: HTML to plain text conversion
- **Repository**: https://github.com/k3a/html2text
- **License**: MIT License
- **License File**: html2text.LICENSE

### p3bot/agentdex (Mozilla Public License 2.0)

- **Purpose**: Agent catalog and skills-directory path resolution
- **Repository**: https://github.com/p3bot/agentdex
- **License**: Mozilla Public License 2.0
- **License File**: agentdex.LICENSE

### go.yaml.in/yaml (MIT and Apache-2.0)

- **Purpose**: YAML unmarshal of skill `SKILL.md` frontmatter
- **Repository**: https://github.com/yaml/go-yaml
- **License**: MIT (ported libyaml files) and Apache License 2.0 (remainder)
- **License File**: go-yaml.LICENSE

## License Compatibility

These licenses are compatible with snag's MPL 2.0:

- MIT and Apache-2.0 are permissive and may be included in an MPL 2.0 project
- agentdex is MPL 2.0, the same license as snag
- Proper attribution is maintained in this directory

## Generating License Files

To verify or update these licenses:

```bash
# Download license files from repositories
curl -L https://raw.githubusercontent.com/go-rod/rod/main/LICENSE -o rod.LICENSE
curl -L https://raw.githubusercontent.com/spf13/cobra/v1.10.2/LICENSE.txt -o cobra.LICENSE
curl -L https://raw.githubusercontent.com/JohannesKaufmann/html-to-markdown/main/LICENSE -o html-to-markdown.LICENSE
curl -L https://raw.githubusercontent.com/k3a/html2text/v1.4.0/LICENSE -o html2text.LICENSE
curl -L https://raw.githubusercontent.com/p3bot/agentdex/v1.1.0/LICENSE -o agentdex.LICENSE
curl -L https://raw.githubusercontent.com/yaml/go-yaml/v3.0.4/LICENSE -o go-yaml.LICENSE
```

## Acknowledgments

We thank the maintainers and contributors of these excellent open-source projects.
