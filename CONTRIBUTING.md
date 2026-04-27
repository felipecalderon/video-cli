CONTRIBUTING.md

Thank you for wanting to contribute to this project!

Contents
- Code of conduct
- Reporting issues
- How to propose changes (Pull Requests)
- Code style and testing
- Commit messages and branches
- Review process

1. Code of conduct
Please see CODE_OF_CONDUCT.md for the project's code of conduct.

2. Reporting an issue
- Search for an existing issue first.
- Include a clear title and a description with:
  - Steps to reproduce
  - Expected and observed behavior
  - Project version and OS
  - Logs or screenshots (if applicable)
- Label the issue when appropriate (bug, enhancement, question).

3. How to propose changes (Pull Request)
- Create a branch from the main branch ("main") using a descriptive name: feature/name, fix/name, docs/name.
- Make atomic commits with clear messages in English or Spanish.
- Open a PR against the "main" branch and describe:
  - What changes and why
  - How to test the changes
  - Reference the related issue (e.g., "Fixes #123")
- A PR should focus on a single purpose. Avoid oversized PRs.

4. Code style and tools
- Follow the repository style. For Go:
  - Run: go fmt ./... and go vet ./...
  - Run tests: go test ./...
- If adding JavaScript/TypeScript or other languages, respect existing tools and configurations (linters, prettier, etc.).

5. Tests
- Add tests covering the changes when relevant.
- Tests should pass locally before creating the PR.

6. Commit messages
- Keep the title under 50 characters, followed by a blank line and a more detailed description if needed.
- Use the imperative mood: "Add feature", "Fix bug".
- Mention related issues: "Fixes #123".

7. Review process
- Maintainers will review the PR. Changes may be requested.
- Address feedback by updating commits or adding new ones as needed.
- Maintainers may rewrite history (rebase/squash) before merging.

8. Signatures and license
By contributing you agree that your contribution will be licensed under the project's license (see LICENSE).

9. Questions and contact
If you have questions, open an issue or contact the maintainers via PR.

Thank you for contributing — your help improves the project.