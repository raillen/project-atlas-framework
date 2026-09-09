#include <iostream>
#include <memory>
#include <span>
#include <vector>

void printNumbers(std::span<const int> nums) {
    for (int n : nums) {
        std::cout << n << '\n';
    }
}

int main() {
    auto numbers = std::make_unique<std::vector<int>>();
    numbers->push_back(1);
    numbers->push_back(2);
    printNumbers(*numbers);
    return 0;
}
