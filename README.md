# openvswitch.patchwork.mgmt

Tools for facilitating the patchwork for Open vSwitch

## What it does

* Report patches that are still marked as **NEW** more than 30 days after post.
* Report duplicated patches (e.g. multiple versions of same patch found).
* Report patches that have been pushed but still marked as **NEW**.

## Install

**1. Follow the links below to install golang and setup the golang directory structure:**

[golang_install](https://golang.org/doc/install)

[golang_dir_setup](https://golang.org/doc/code.html)

**2. Clone the repo into go/src/github.com/yew011/**

**3. Execute `go get && go build`, the excutable will be generated at ./openvswitch.patchwork.mgmt**

## Run
	alex@debian-jessie:~/alex_dev/go/src/github.com/yew011/openvswitch.patchwork.mgmt$ ./openvswitch.patchwork.mgmt --help
	NAME:
	   ovs-patchwork - Tool for ovs patchwork facilitation.  User must provide the
	   				   --ovs-dir and --ovs-commit options.

	USAGE:
	   ovs-patchwork [global options] command [command options] [arguments...]

	VERSION:
	   0.0.0

	COMMANDS:
	   help, h      Shows a list of commands or help for one command

	GLOBAL OPTIONS:
	   --ovs-dir            Path to ovs git repo
	   --ovs-commit         Commit to start check for committed patches
	   --mark-committed     Mark the committed patch as 'Accepted'
	   --mark-dup           Mark the duplicate patch as 'Not Applicable'
	   --help, -h           show help
	   --version, -v        print the version

## Example Output

    30+ Day Old Patches
    ===================
    ID      State  Date                  Name
    --      -----  ----                  ----
    516388  New    2015-09-10 18:54:25   [ovs-dev] ovn-nb: Add port_security proposal.
    516433  New    2015-09-10 20:18:55   [ovs-dev,2/2] ofproto: Correctly reject duplicate bucket ID for OFPGC_INSERT_BUCKET.
    519017  New    2015-09-17 20:29:45   [ovs-dev] dpif-netdev: move header prefetch earlier into the receive function

    Duplicate Patches in Patchwork
    ==============================
    ID      State  Date                  Name
    --      -----  ----                  ----
    539664  New    2015-11-04 00:38:11   [ovs-dev,01/11] ct-dpif: New module.
    539666  New    2015-11-04 00:38:12   [ovs-dev,02/11] netlink-conntrack: New module.
    539657  New    2015-11-04 00:38:13   [ovs-dev,03/11] ct-dpif: Add ct_dpif_dump_{start, next, done}().
    539665  New    2015-11-04 00:38:14   [ovs-dev,04/11] ct-dpif: Add ct_dpif_flush().
    539663  New    2015-11-04 00:38:15   [ovs-dev,05/11] dpif-netlink: Implement ct_dump_{start, next, done}.
    539662  New    2015-11-04 00:38:16   [ovs-dev,06/11] dpctl: Add 'conntrack-dump' command.
    539656  New    2015-11-04 00:38:17   [ovs-dev,07/11] dpif-netlink: Implement ct_flush.
    539658  New    2015-11-04 00:38:18   [ovs-dev,08/11] dpctl: Add new 'flush-conntrack' command.
    539660  New    2015-11-04 00:38:19   [ovs-dev,09/11] ovs-test: Add test-netlink-conntrack command.
    539661  New    2015-11-04 00:38:20   [ovs-dev,10/11] system-traffic: use `dpctl/*conntrack` instead of `conntrack` tool.
    539659  New    2015-11-04 00:38:21   [ovs-dev,11/11] system-kmod-macros: Do not require the 'conntrack' tool.
