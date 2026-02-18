cmake_minimum_required(VERSION 4.1)
project(tau)

set(CMAKE_CXX_STANDARD 20)

find_package(LibLZMA REQUIRED)

add_library(tau-lib SHARED
        src/compression.cpp
        src/compression.h
        src/packing.cpp
        src/packing.h
)

add_executable(tau src/main.cpp)


target_link_libraries(tau-lib PRIVATE LibLZMA::LibLZMA)
target_link_libraries(tau PRIVATE tau-lib)