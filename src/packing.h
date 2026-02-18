#ifndef TAU_PACKING_H
#define TAU_PACKING_H
#include <string>

#include "compression.h"

class packer {
public:
	packer (const std::string &dir, const std::string &name);
	~packer ();
	void add_packing (const std::string &dir);

private:
	std::string dir;
	compression_list *clist;
};

#endif // TAU_PACKING_H
