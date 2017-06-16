property :user_name, String, name_property: true

property :binary_path, String, default: '/usr/local/bin/mc'
property :binary_url, kind_of: [String, NilClass], default: nil

property :server_search_query, String, default: 'minio_server_name:*'
property :add_hosts, kind_of: [Array, Symbol], default: nil
property :s3_access_key, kind_of: [String, NilClass], default: nil
property :s3_secret_key, kind_of: [String, NilClass], default: nil

default_action [:install, :configure]

action :install do
  remote_file binary_path do
    source MinioURLHelper.get('mc', node, binary_url)
    owner 'root'
    mode '755'
    action :create_if_missing
  end
end

action :configure do
  unless add_hosts && add_hosts.empty?
    servers = search( 'node',
                    server_search_query,
                    :filter_result => { 'server' => ['minio', 'server'] })
    if add_hosts
      servers.reject! { |s| !add_hosts.include? s['server']['name'] }
    end

    servers.each do |s|
      s = s['server']
      execute "set-up mc server #{s['name']}" do
        scheme = s['ssl'] ? 'https://' : 'http://'
        url = s['address']
        url[/^(?:https?:\/\/)?/] = scheme

        command "mc config host remove #{s['name']} && " +
                "mc config host add #{s['name']} #{url} #{s['access_key']} #{s['secret_key']}"
        user user_name
        environment({ 'HOME' => ::Dir.home(user_name), 'USER' => user_name })
      end
    end
  end

  if s3_access_key && s3_secret_key
    execute 'set-up mc s3' do
      command "mc config host remove s3 && " + "mc config host add s3 https://s3.amazonaws.com #{s3_access_key} #{s3_secret_key}"
      user user_name
      environment({ 'HOME' => ::Dir.home(user_name), 'USER' => user_name })
    end
  end
end
