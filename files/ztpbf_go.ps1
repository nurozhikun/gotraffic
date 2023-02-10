# C:/protobuf/bin/protoc.exe -I"./protbee/" --go_out="../../zk-atom/zipbf/protbee" "./protbee/prot_bee_base.proto" 
$name = "C:/tools/protoc-21.12-win64/bin/protoc.exe "
$name += ' -I"./ztpbf/"'
$name += ' --go_out="../ztpbf/"'
$name += ' "./ztpbf/traffic.proto"'
echo $name
cmd /C $name
#  --dart_out="../../../sieflutter/fzksie/lib/pbf" 
# --cpp_out="../../../../geminicpp/src/pbf/robhost"
# --dart_out="../../../sieflutter/fzksie/lib/pbf"