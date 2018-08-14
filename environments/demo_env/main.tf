provider "aws" {
  region = "${var.region}"
}

data "aws_availability_zones" "available_azs" {}

resource "aws_vpc" "vpc" {
  cidr_block = "${var.cidr_block}"

  tags {
    Name = "${var.vpc_name}"
    ManagedBy = "Terraform"
  }
}

resource "aws_internet_gateway" "igw" {
  vpc_id = "${aws_vpc.vpc.id}"

  tags {
    Name = "${var.vpc_name}"
    ManagedBy = "Terraform"
  }
}

resource "aws_route_table" "rt" {
  vpc_id = "${aws_vpc.vpc.id}"
}

resource "aws_route" "outbound" {
  route_table_id = "${aws_route_table.rt.id}"
  destination_cidr_block = "0.0.0.0/0"
  gateway_id = "${aws_internet_gateway.igw.id}"
}

resource "aws_subnet" "subnets" {
  count = "${length(data.aws_availability_zones.available_azs.names)}"
  vpc_id = "${aws_vpc.vpc.id}"
  cidr_block = "${cidrsubnet(aws_vpc.vpc.cidr_block, 8, count.index)}"
}

resource "aws_route_table_association" "rta" {
  count = "${length(data.aws_availability_zones.available_azs.names)}"
  route_table_id = "${aws_route_table.rt.id}"
  subnet_id = "${element(aws_subnet.subnets.*.id, count.index)}"
}

data "aws_ami" "amazon_linux" {
  most_recent = true

  filter {
    name = "owner-alias"
    values = ["amazon"]
  }

  filter {
    name = "name"
    values = ["amzn2-ami-hvm*"]
  }
}

resource "aws_security_group" "sg" {
  vpc_id = "${aws_vpc.vpc.id}"
  name = "test_instance"
}

resource "aws_security_group_rule" "outbound" {
  security_group_id = "${aws_security_group.sg.id}"
  type = "egress"
  protocol = -1
  from_port = 0
  to_port = 65535
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_security_group_rule" "ssh" {
  security_group_id = "${aws_security_group.sg.id}"
  type = "ingress"
  protocol = "tcp"
  from_port = 22
  to_port = 22
  cidr_blocks = ["0.0.0.0/0"]
}

resource "aws_instance" "test_instance" {
  ami = "${data.aws_ami.amazon_linux.id}"
  instance_type = "t2.micro"
  key_name = "${var.ssh_key_name}"
  subnet_id = "${aws_subnet.subnets.0.id}"
  vpc_security_group_ids = ["${aws_security_group.sg.id}"]
  associate_public_ip_address = true

  tags {
    Name = "${var.vpc_name}-test-instance"
    ManagedBy = "Terraform"
  }
}
