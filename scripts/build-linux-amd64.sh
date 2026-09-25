#!/usr/bin/env bash
set -Eeuo pipefail

script_dir=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
primary_root=$(cd -- "$script_dir/.." && pwd)
expected_root=$(git -C "$primary_root" rev-parse --show-toplevel)
if [[ "$expected_root" != "$primary_root" ]]; then
	printf 'error: script root is not the Primary git root: %s\n' "$primary_root" >&2
	exit 1
fi

primary_branch=$(git -C "$primary_root" branch --show-current)
if [[ "$primary_branch" != main ]]; then
	printf 'error: expected Primary branch main, found %s\n' "${primary_branch:-detached HEAD}" >&2
	exit 1
fi
primary_git_dir=$(git -C "$primary_root" rev-parse --absolute-git-dir)
if git -C "$primary_root" rev-parse -q --verify MERGE_HEAD >/dev/null 2>&1 || \
	git -C "$primary_root" rev-parse -q --verify CHERRY_PICK_HEAD >/dev/null 2>&1 || \
	git -C "$primary_root" rev-parse -q --verify REVERT_HEAD >/dev/null 2>&1 || \
	[[ -d "$primary_git_dir/rebase-merge" ]] || [[ -d "$primary_git_dir/rebase-apply" ]] || \
	[[ -d "$primary_git_dir/sequencer" ]]; then
	printf 'error: Primary has an in-progress Git operation\n' >&2
	exit 1
fi
if [[ -n "$(git -C "$primary_root" diff --name-only --diff-filter=U)" ]]; then
	printf 'error: Primary has unresolved index conflicts\n' >&2
	exit 1
fi

backend_root=$(cd -- "$primary_root/../lattigo" 2>/dev/null && pwd) || {
	printf 'error: Secondary repository ../lattigo does not exist\n' >&2
	exit 1
}
if [[ "$(git -C "$backend_root" rev-parse --is-inside-work-tree 2>/dev/null)" != true ]]; then
	printf 'error: Secondary path is not a Git worktree: %s\n' "$backend_root" >&2
	exit 1
fi
if [[ -n "$(git -C "$backend_root" status --porcelain --untracked-files=all)" ]]; then
	printf 'error: Secondary worktree is dirty; refusing to switch branches\n' >&2
	exit 1
fi

original_branch=$(git -C "$backend_root" branch --show-current)
original_commit=$(git -C "$backend_root" rev-parse HEAD)
if [[ -n "$original_branch" && "$(git -C "$backend_root" rev-parse "refs/heads/$original_branch")" != "$original_commit" ]]; then
	printf 'error: Secondary HEAD is not at the tip of branch %s\n' "$original_branch" >&2
	exit 1
fi
for branch in main fast-ckks; do
	if ! git -C "$backend_root" show-ref --verify --quiet "refs/heads/$branch"; then
		printf 'error: Secondary local branch %s is missing\n' "$branch" >&2
		exit 1
	fi
done

secondary_switch_started=0
standard_tmp=''
fast_tmp=''

restore_secondary() {
	local status
	if [[ "$secondary_switch_started" != 1 ]]; then
		return 0
	fi
	if [[ -n "$(git -C "$backend_root" status --porcelain --untracked-files=all)" ]]; then
		printf 'error: Secondary became dirty; refusing to switch away from user changes\n' >&2
		return 1
	fi
	if [[ -n "$original_branch" ]]; then
		git -C "$backend_root" switch "$original_branch" || return 1
	else
		git -C "$backend_root" switch --detach "$original_commit" || return 1
	fi
	if [[ "$(git -C "$backend_root" rev-parse HEAD)" != "$original_commit" ]]; then
		printf 'error: Secondary returned to a different commit than its original HEAD\n' >&2
		return 1
	fi
	status=$(git -C "$backend_root" status --porcelain --untracked-files=all)
	if [[ -n "$status" ]]; then
		printf 'error: Secondary is not clean after restoration\n' >&2
		return 1
	fi
}

on_exit() {
	local result=$?
	trap - EXIT
	if [[ -n "$standard_tmp" ]]; then
		rm -f -- "$standard_tmp"
	fi
	if [[ -n "$fast_tmp" ]]; then
		rm -f -- "$fast_tmp"
	fi
	if ! restore_secondary; then
		printf 'error: failed to restore Secondary branch/commit; original was %s at %s\n' \
			"${original_branch:-detached HEAD}" "$original_commit" >&2
		if [[ "$result" == 0 ]]; then
			result=1
		fi
	fi
	exit "$result"
}
trap on_exit EXIT
trap 'exit 130' INT
trap 'exit 143' TERM

dist_dir="$primary_root/dist"
mkdir -p -- "$dist_dir"
standard_tmp="$dist_dir/.bootstrap-standard-linux-amd64.$$"
fast_tmp="$dist_dir/.bootstrap-fast-linux-amd64.$$"

secondary_switch_started=1
git -C "$backend_root" switch main
(
	cd -- "$primary_root"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
		go build -trimpath -tags lattigo_standard -o "$standard_tmp" .
)
mv -- "$standard_tmp" "$dist_dir/bootstrap-standard-linux-amd64"
standard_tmp=''

git -C "$backend_root" switch fast-ckks
(
	cd -- "$primary_root"
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GOAMD64=v1 \
		go build -trimpath -o "$fast_tmp" .
)
mv -- "$fast_tmp" "$dist_dir/bootstrap-fast-linux-amd64"
fast_tmp=''

printf 'built %s\n' "$dist_dir/bootstrap-standard-linux-amd64"
printf 'built %s\n' "$dist_dir/bootstrap-fast-linux-amd64"
