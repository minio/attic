property :data_dir, name_property: true
property :server_name, String, default: node.name
property :address, String, default: ":9000"
property :binary_path, default: '/usr/local/sbin/minio'
property :binary_url, kind_of: [String, nil], default: nil

property :make_user, kind_of: [TrueClass, FalseClass], default: true
property :user_name, default: 'minio'
property :user_home, default: lazy { data_dir }

property :service_name, String, default: 'minio'
property :service_provider, Symbol, default: :sysvinit

property :publish_info, kind_of: [TrueClass, FalseClass], default: lazy { !Chef::Config[:solo] && !Chef::Config.local_mode }


action :create do
  poise_service_user user_name do
    home user_home
    action :create
    only_if { make_user }
  end

  directory data_dir do
    owner lazy { user_name }
  end

  remote_file binary_path do
    source MinioURLHelper.get('minio', node, binary_url)
    owner user_name
    mode '755'
    action :create_if_missing
  end

  poise_service service_name do
    provider service_provider
    command "#{binary_path} server --address #{address} #{data_dir}"
    user user_name
    action [:enable, :start]
  end

  ruby_block 'publish minio info' do
    block do
      require 'json'
      j = JSON.parse(::File.read(::File.join(user_home, '.minio', 'config.json')))
      node.set['minio']['server'].tap do |m|
        m['name'] = server_name
        m['ssl'] = false
        m['address'] = address !~ /^:/ ? address : node.ipaddress + address
        m['access_key'] = j["credentials"]["accessKeyId"]
        m['secret_key'] = j["credentials"]["secretAccessKey"]
        m['region'] = j["credentials"]["region"]
      end
    end

    only_if { publish_info }
  end
end
