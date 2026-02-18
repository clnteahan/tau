#include "packing.h"

packing_list::packing_list(const std::string &dir, const std::string& name) {
	this->dir = dir;
	this->clist = new compression_list(dir + "/" + name + ".tar.xz");
}

packing_list::~packing_list() {
	delete this->clist;
}
