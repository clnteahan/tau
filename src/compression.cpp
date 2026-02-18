#include <lzma.h>
#include <filesystem>
#include <fstream>
#include <vector>
#include <string>
#include <iostream>
#include "compression.h"

namespace fs = std::filesystem;

#ifndef TAU_LZMA_BUFFER_SIZE
	#define TAU_LZMA_BUFFER_SIZE 8192
#endif

// Compress a single file using LZMA
bool compressFile(const std::string& inputPath, const std::string& outputPath) {
    // Open input file
    std::ifstream inFile(inputPath, std::ios::binary);
    if (!inFile) {
        std::cerr << "Failed to open input file: " << inputPath << std::endl;
        return false;
    }

    // Open output file
    std::ofstream outFile(outputPath, std::ios::binary);
    if (!outFile) {
        std::cerr << "Failed to open output file: " << outputPath << std::endl;
        return false;
    }

    // Initialize LZMA encoder
    lzma_stream strm = LZMA_STREAM_INIT;
    lzma_ret ret = lzma_easy_encoder(&strm, 6, LZMA_CHECK_CRC64);
    
    if (ret != LZMA_OK) {
        std::cerr << "Failed to initialize LZMA encoder" << std::endl;
        return false;
    }

    // Buffer sizes
    std::vector<uint8_t> inBuffer(TAU_LZMA_BUFFER_SIZE);
    std::vector<uint8_t> outBuffer(TAU_LZMA_BUFFER_SIZE);

    strm.next_in = nullptr;
    strm.avail_in = 0;
    strm.next_out = outBuffer.data();
    strm.avail_out = TAU_LZMA_BUFFER_SIZE;

    lzma_action action = LZMA_RUN;
    bool success = true;

    while (true) {
        // Read input if buffer is empty
        if (strm.avail_in == 0 && !inFile.eof()) {
            inFile.read(reinterpret_cast<char*>(inBuffer.data()), TAU_LZMA_BUFFER_SIZE);
            strm.next_in = inBuffer.data();
            strm.avail_in = inFile.gcount();

            if (inFile.eof()) {
                action = LZMA_FINISH;
            }
        }

        // Compress
        ret = lzma_code(&strm, action);

        // Write output if buffer is full or compression is done
        if (strm.avail_out == 0 || ret == LZMA_STREAM_END) {
            size_t writeSize = TAU_LZMA_BUFFER_SIZE - strm.avail_out;
            outFile.write(reinterpret_cast<char*>(outBuffer.data()), writeSize);
            
            if (!outFile) {
                std::cerr << "Failed to write to output file" << std::endl;
                success = false;
                break;
            }

            strm.next_out = outBuffer.data();
            strm.avail_out = TAU_LZMA_BUFFER_SIZE;
        }

        if (ret != LZMA_OK) {
            if (ret == LZMA_STREAM_END) {
                break;
            }
            std::cerr << "LZMA encoding error: " << ret << std::endl;
            success = false;
            break;
        }
    }

    lzma_end(&strm);
	inFile.close();
	outFile.close();
    return success;
}

// Recursively compress a directory to a tar-like format with LZMA
bool compressDirectory(const std::string& dirPath, const std::string& outputPath) {
    // Create a temporary file list
    std::vector<fs::path> files;

    try {
        for (const auto& entry : fs::recursive_directory_iterator(dirPath)) {
            if (entry.is_regular_file()) {
                files.push_back(entry.path());
            }
        }
    } catch (const fs::filesystem_error& e) {
        std::cerr << "Error accessing directory: " << e.what() << std::endl;
        return false;
    }

    // Open output file
    std::ofstream outFile(outputPath, std::ios::binary);
    if (!outFile) {
        std::cerr << "Failed to open output file: " << outputPath << std::endl;
        return false;
    }

    // Initialize LZMA encoder
    lzma_stream strm = LZMA_STREAM_INIT;
    lzma_ret ret = lzma_easy_encoder(&strm, 6, LZMA_CHECK_CRC64);

    if (ret != LZMA_OK) {
        std::cerr << "Failed to initialize LZMA encoder" << std::endl;
        return false;
    }

    std::vector<uint8_t> inBuffer(TAU_LZMA_BUFFER_SIZE);
    std::vector<uint8_t> outBuffer(TAU_LZMA_BUFFER_SIZE);

    strm.next_out = outBuffer.data();
    strm.avail_out = TAU_LZMA_BUFFER_SIZE;

    auto writeCompressed = [&](const uint8_t* data, size_t size, lzma_action action) {
        strm.next_in = data;
        strm.avail_in = size;

        while (strm.avail_in > 0 || action == LZMA_FINISH) {
            ret = lzma_code(&strm, action);

            if (strm.avail_out == 0 || ret == LZMA_STREAM_END) {
                size_t writeSize = TAU_LZMA_BUFFER_SIZE - strm.avail_out;
                outFile.write(reinterpret_cast<char*>(outBuffer.data()), writeSize);
                strm.next_out = outBuffer.data();
                strm.avail_out = TAU_LZMA_BUFFER_SIZE;
            }

            if (ret == LZMA_STREAM_END || (action == LZMA_RUN && strm.avail_in == 0)) {
                break;
            }
        }
        return ret == LZMA_OK || ret == LZMA_STREAM_END;
    };

    // Compress each file with metadata
    fs::path basePath(dirPath);
    for (size_t i = 0; i < files.size(); ++i) {
        const auto& filePath = files[i];
        std::string relativePath = fs::relative(filePath, basePath).string();

        // Write file metadata (path length, path, file size)
        uint32_t pathLen = relativePath.size();
        writeCompressed(reinterpret_cast<uint8_t*>(&pathLen), sizeof(pathLen), LZMA_RUN);
        writeCompressed(reinterpret_cast<const uint8_t*>(relativePath.c_str()), pathLen, LZMA_RUN);

        uint64_t fileSize = fs::file_size(filePath);
        writeCompressed(reinterpret_cast<uint8_t*>(&fileSize), sizeof(fileSize), LZMA_RUN);

        // Write file content
        std::ifstream inFile(filePath, std::ios::binary);
        while (inFile) {
            inFile.read(reinterpret_cast<char*>(inBuffer.data()), TAU_LZMA_BUFFER_SIZE);
            size_t bytesRead = inFile.gcount();
            if (bytesRead > 0) {
                writeCompressed(inBuffer.data(), bytesRead, LZMA_RUN);
            }
        }
    }

    // Finish compression
    writeCompressed(nullptr, 0, LZMA_FINISH);

    lzma_end(&strm);
    return true;
}

// Main function to compress file or directory
bool compressPath(const std::string& inputPath, const std::string& outputPath) {
    if (!fs::exists(inputPath)) {
        std::cerr << "Input path does not exist: " << inputPath << std::endl;
        return false;
    }

    if (fs::is_directory(inputPath)) {
        std::cout << "Compressing directory: " << inputPath << std::endl;
        return compressDirectory(inputPath, outputPath);
    } else if (fs::is_regular_file(inputPath)) {
        std::cout << "Compressing file: " << inputPath << std::endl;
        return compressFile(inputPath, outputPath);
    } else {
        std::cerr << "Input path is neither a file nor a directory" << std::endl;
        return false;
    }
}

compression_list::compression_list(const std::string &outputPath) {
	this->oPath = outputPath;
}

compression_list::~compression_list() {
	for (auto iter = this->list.begin(); iter != this->list.end(); ++iter) {
		delete *iter;
	}
	this->list.clear();
}

void compression_list::add(const std::string &inputPath) {
	if (!fs::exists(inputPath)) {
		std::cerr << "Input path does not exist: " << inputPath << std::endl;
	}
	const auto cpy = new std::string(inputPath);

	this->list.push_back(cpy);
}



bool compression_list::compress() const {

	for (int i = 0; i < this->list.size(); i++) {
		//if ()
	}
	return false;
}