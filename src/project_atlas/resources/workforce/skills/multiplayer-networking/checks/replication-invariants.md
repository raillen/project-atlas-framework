# Replication Invariants Checklist

- [ ] Server validates all inputs before applying state changes.
- [ ] Client prediction limits deviation to maximum 100ms equivalent.
- [ ] Delta compression does not omit critical unacknowledged state.
- [ ] No remote code execution vulnerabilities in packet parsing.
