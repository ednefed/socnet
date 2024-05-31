#!/usr/bin/env sh

# Telegraf usage example
#
# [[inputs.exec]]
#   name_override = "net"
#   tag_keys = ["interface"]
#   commands = ["sh /net.sh"]
#   data_format = "influx"

function main() {
  hostfs="${HOST_MOUNT_PREFIX}"

  # Exclude localhost, virtual and docker-related interfaces
  interfaces=$(ls "${hostfs}/sys/class/net/" | grep -vE "^(lo|br-|docker|veth)")

  for iface in ${interfaces}
  do
    iface_stats="${hostfs}/sys/class/net/${iface}/statistics"

    bytes_sent=$(cat "${iface_stats}/tx_bytes")
    bytes_recv=$(cat "${iface_stats}/rx_bytes")
    packets_sent=$(cat "${iface_stats}/tx_packets")
    packets_recv=$(cat "${iface_stats}/rx_packets")
    err_out=$(cat "${iface_stats}/tx_errors")
    err_in=$(cat "${iface_stats}/rx_errors")
    drop_out=$(cat "${iface_stats}/tx_dropped")
    drop_in=$(cat "${iface_stats}/rx_dropped")

    if [ -n "${bytes_sent}" ] &&
      [ -n "${bytes_recv}" ] &&
      [ -n "${packets_sent}" ] &&
      [ -n "${packets_recv}" ] &&
      [ -n "${err_in}" ] &&
      [ -n "${err_out}" ] &&
      [ -n "${drop_in}" ] &&
      [ -n "${drop_out}" ]
    then
      # Use same fields as present in inputs.net plugin
      influxdb=$(echo "${influxdb}net,interface=${iface} bytes_sent=${bytes_sent}i,bytes_recv=${bytes_recv}i,packets_sent=${packets_sent}i,packets_recv=${packets_recv}i,err_in=${err_in}i,err_out=${err_out}i,drop_in=${drop_in}i,drop_out=${drop_out}i\n")
    fi
  done

  # Remove trailing newline
  influxdb="$(echo "${influxdb}" | sed 's/\\n$//')"

  echo -e "${influxdb}"
}

main $@; exit $?
