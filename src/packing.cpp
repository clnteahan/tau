#include "packing.h"

packer::packer(const std::string &dir, const std::string& name) {
	this->dir = dir;
	this->clist = new compression_list(dir + "/" + name + ".tar.xz");
}

packer::~packer() {
	delete this->clist;
}
