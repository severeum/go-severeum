Pod::Spec.new do |spec|
  spec.name         = 'Seth'
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
	spec.ios.vendored_frameworks = 'Frameworks/Seth.framework'

	spec.prepare_command = <<-CMD
    curl https://sethstore.blob.core.windows.net/builds/{{.Archive}}.tar.gz | tar -xvz
    mkdir Frameworks
    mv {{.Archive}}/Seth.framework Frameworks
    rm -rf {{.Archive}}
  CMD
end
