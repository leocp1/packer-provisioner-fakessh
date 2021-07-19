{ stdenv
, lib
, buildGoModule
, gitignoreSource
, nixFilter
}:
buildGoModule rec {
  pname = "packer-provisioner-fakessh";
  version = "0.0.1";
  src = nixFilter (gitignoreSource ./.);
  vendorSha256 = "1hafkq1djhfiw2f3gh36y1fhq6jknqy7qf5clbb841x2gws9kn0q";
  doCheck = true;
  patchPhase = ''
    substituteAllInPlace ./pkg/fakessh/utils.go
  '';
  postFixup = ''
    install -Dm755 $out/bin/ssh $out/share/bin/ssh
    rm $out/bin/ssh
  '';
  meta = with stdenv.lib; {
    description = "Packer provisioner with fake ssh command";
    license = licenses.mpl20;
    maintainers = with maintainers; [ leocp1 ];
    platforms = platforms.linux;
  };
}
