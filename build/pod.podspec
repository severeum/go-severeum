Pod::Spec.new do |spec|
  spec.name         = 'Ssev'
  spec.version      = '{{.Version}}'
  spec.license      = { :type => 'GNU Lesser General Public License, Version 3.0' }
  spec.homepage     = 'https://github.com/severeum/go-severeum'
  spec.authors      = { {{range .Contributors}}
		'{{.Name}}' => '{{.Email}}',{{end}}
	}
  spec.summary      = 'iOS Severeum Client'
  spec.source       = { :git => 'https://github.com/severeum/go-severeum.git', :commit => '{{.Commit}}' }

	spec.platform = :ios
  spec.ios.deployment_target  = '9.0'
	spec.ios.vendored_frameworks = 'Frameworks/Ssev.framework'

	spec.prepare_command = <<-CMD
    curl https://ssevstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Ssev.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
