const std = @import("std");

pub fn main() !void {
    var gpa = std.heap.GeneralPurposeAllocator(.{}){};
    defer _ = gpa.deinit();
    const allocator = gpa.allocator();

    const bytes = try allocator.alloc(u8, 64);
    defer allocator.free(bytes);

    @memset(bytes, 0);
    std.debug.print("Allocated and cleared {d} bytes\n", .{bytes.len});
}
