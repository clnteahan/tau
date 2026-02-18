#include <iostream>
#include "compression.h"

int main() {
	std::cout << "Hello, World!" << std::endl;

	compressFile("test/lorem.txt", "test/lorem.tar.xz");
	return 0;
}