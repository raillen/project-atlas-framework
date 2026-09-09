#include <iostream>

void mmio_access() {
    int* hw = reinterpret_cast<int*>(0x40001000);
    (void)hw;
}

int main() {
    mmio_access();
    return 0;
}
