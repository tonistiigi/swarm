#!/usr/bin/env bats

load ../helpers

function teardown() {
	swarm_manage_cleanup
	stop_docker
}

@test "docker inspect" {
	local version="new"
	start_docker_with_busybox 2
	swarm_manage
	# run container
	docker_swarm run -d -e TEST=true -h hostname.test --name test_container busybox sleep 500

	# make sure container exists
	run docker_swarm ps -l
	[ "${#lines[@]}" -eq 2 ]
	[[ "${lines[1]}" == *"test_container"* ]]

	# inspect and verify
	run docker_swarm inspect test_container
	[ "$status" -eq 0 ]
	[[ "${output}" == *"NetworkSettings"* ]]
	[[ "${output}" == *"TEST=true"* ]]
	# the specific information of swarm node
	[[ "${output}" == *'"Node": {'* ]]
	[[ "${output}" == *'"Name": "node-'* ]]
	[[ "${output}" == *'"Hostname": "hostname.test"'* ]]
	[[ "${output}" == *'"Domainname": ""'* ]]
}

@test "docker inspect --format" {
	start_docker_with_busybox 2
	swarm_manage
	# run container
	docker_swarm run -d --name test_container busybox sleep 500

	# make sure container exists
	run docker_swarm ps -l
	[ "${#lines[@]}" -eq 2 ]
	[[ "${lines[1]}" == *"test_container"* ]]

	# inspect --format='{{.Config.Image}}', return one line: image name
	run docker_swarm inspect --format='{{.Config.Image}}' test_container
	[ "$status" -eq 0 ]
	[ "${#lines[@]}" -eq 1 ]
	[[ "${lines[0]}" == "busybox" ]]

	# inspect --format='{{.Node.IP}}', return one line: Node ip
	run docker_swarm inspect --format='{{.Node.IP}}' test_container
	[ "$status" -eq 0 ]
	[ "${#lines[@]}" -eq 1 ]
	[[ "${lines[0]}" == "127.0.0.1" ]]
}
