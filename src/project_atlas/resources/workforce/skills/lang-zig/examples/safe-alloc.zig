const std = @import("std");

pub fn allocateBuffer(allocator: std.mem.Allocator, size: usize) ![]u8 {
    const buf = try allocator.alloc(u8, size);
    errdefer allocator.free(buf);
    @memset(buf, 0);
    return buf;
}

test "allocateBuffer leak test" {
    const testing_allocator = std.testing.allocator;
    const buf = try allocateBuffer(testing_allocator, 128);
    defer testing_allocator.free(buf);
    try std.testing.expectEqual(buf.len, 128);
}
