# Look up image by ID (get image_id from STACKIT console/API or set via variable)
data "stackit_image" "ubuntu_24_04_kevin" {
	project_id = var.project_id
	image_id   = var.image_id
}

resource "stackit_volume" "SKE_volume_kevin" {
	project_id			=	var.project_id
	name				=	"SKE_volume_kevin"
	availability_zone	=	"eu01-1"
	description			=	"Volume for SKE cluster"
	size				=	64
	performance_class	=	"storage_premium_perf2"
	source = {
		id   = data.stackit_image.ubuntu_24_04_kevin.image_id
		type = "image"
	}
	depends_on = [data.stackit_image.ubuntu_24_04_kevin]
}
