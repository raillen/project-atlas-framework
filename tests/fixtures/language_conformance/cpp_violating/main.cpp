#include <iostream>
#include <cstdlib>

void badFunction() {
    int* p = reinterpret_cast<int*>(0xdeadbeef);
    void* v = (void*)p;
    int* leak = (int*)malloc(sizeof(int));
    goto finish;
finish:
    return;
}

int main() {
    badFunction();
    return 0;
}
