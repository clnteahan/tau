#ifndef TAU_COMPRESSION_H
#define TAU_COMPRESSION_H

#include <vector>

// Compress a single file using LZMA
bool compressFile(const std::string& inputPath, const std::string& outputPath);

// Recursively compress a directory to a tar-like format with LZMA
bool compressDirectory(const std::string& dirPath, const std::string& outputPath);

// Main function to compress file or directory
bool compressPath(const std::string& inputPath, const std::string& outputPath);

class compression_list {
public:
	explicit compression_list(const std::string& outputPath);
	~compression_list();
	void add(const std::string& inputPath);
	bool compress() const;
private:
	std::vector<const std::string*> list;
	std::string oPath;
};

#endif //TAU_COMPRESSION_H