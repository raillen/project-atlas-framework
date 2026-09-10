```rust
trait Visitor {
    fn visit_expr(&mut self, expr: &Expr);
    fn visit_stmt(&mut self, stmt: &Stmt);
}
```
