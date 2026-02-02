data "stackit_image" "ubuntu_24_04" {
	project_id	= var.project_id
	name 		= "Ubuntu 24.04"
}

resource "stackit_volume" "SKE_volume" {
	project_id			=	var.project_id
	name				=	"SKE_volume"
	availability_zone	=	"eu01-1"
	description			=	"Volume for SKE cluster"
	size				=	64
	performance_class	=	"storage_premium_perf2"
	source = {
		id		= data.stackit_image.ubuntu_24_04.image_id
		type	= "image"
	}
	depends_on = [data.stackit_image.ubuntu_24_04]
}
