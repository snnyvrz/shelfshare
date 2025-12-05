#!/usr/bin/env bash
set -euo pipefail

COMMAND="${1:-}"
shift || true

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
TF_DIR="${PROJECT_ROOT}/infra/hetzner"
SOPS_SCRIPT="${PROJECT_ROOT}/scripts/sops.sh"

DECRYPTED_VARS_FILE="secrets.sops.tfvars"

error() {
    echo "Error: $*" >&2
    exit 1
}

get_my_ip() {
    curl -4 -s https://ifconfig.me || curl -4 -s https://api.ipify.org
}

decrypt_tfvars() {
    (
        cd "${TF_DIR}"
        "${SOPS_SCRIPT}" decrypt >&2
    )

    local path="${TF_DIR}/${DECRYPTED_VARS_FILE}"
    [[ -f "${path}" ]] || error "Expected decrypted tfvars at '${path}' but it does not exist."
    echo "${path}"
}

run_in_tf_dir() {
    (cd "${TF_DIR}" && "$@")
}

case "${COMMAND}" in
    init)
        echo "[terraform.sh] Running terraform init in ${TF_DIR}"
        run_in_tf_dir terraform init "$@"
        ;;

    plan)
        echo "[terraform.sh] Determining public IP..."
        MY_IP="$(get_my_ip)"
        [[ -n "${MY_IP}" ]] || error "Could not determine public IP"
        echo "[terraform.sh] Detected public IP: ${MY_IP}"

        VAR_FILE_PATH="$(decrypt_tfvars)"
        echo "[terraform.sh] Using var-file: ${VAR_FILE_PATH}"

        echo "[terraform.sh] Running terraform plan in ${TF_DIR}"
        run_in_tf_dir terraform plan \
            -var="my_home_ip=${MY_IP}" \
            -var-file="${VAR_FILE_PATH}" \
            "$@"
        ;;

    apply)
        echo "[terraform.sh] Determining public IP..."
        MY_IP="$(get_my_ip)"
        [[ -n "${MY_IP}" ]] || error "Could not determine public IP"
        echo "[terraform.sh] Detected public IP: ${MY_IP}"

        VAR_FILE_PATH="$(decrypt_tfvars)"
        echo "[terraform.sh] Using var-file: ${VAR_FILE_PATH}"

        echo "[terraform.sh] Running terraform apply in ${TF_DIR}"
        run_in_tf_dir terraform apply \
            -var="my_home_ip=${MY_IP}" \
            -var-file="${VAR_FILE_PATH}" \
            "$@"
        ;;

    *)
        cat >&2 <<EOF
Usage: $0 <command> [terraform-args...]

Commands:
  init        Run 'terraform init' in infra/hetzner
  plan        Run 'terraform plan' with:
                - var my_home_ip set to your current public IP
                - var-file from decrypted secrets.sops.tfvars (secrets.sops.tfvars)
  apply       Run 'terraform apply' with:
                - var my_home_ip set to your current public IP
                - var-file from decrypted secrets.sops.tfvars (secrets.sops.tfvars)

Examples:
  $0 init
  $0 plan
  $0 apply
EOF
        exit 1
        ;;
esac
