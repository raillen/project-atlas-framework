module example;

import std::io;

fn void processItems(int[] items) {
    io::printfln("Processing %d items safely", items.len);
    foreach (item : items) {
        io::printfln("Item: %d", item);
    }
}

fn int main() {
    int[3] data = { 10, 20, 30 };
    processItems(&data);
    return 0;
}
