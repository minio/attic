class MinioURLHelper
  def self.get(component, node, override_url = nil)
    if override_url
      return override_url
    end

    dir = case component
          when 'minio' then 'server/minio'
          when 'mc' then 'client/mc'
          else
            raise ArgumentError, "unkown minio component: #{component}"
          end

    arch = case node['kernel']['machine']
           when 'x86_64' then 'amd64'
           when 'i686' then '386'
           when /arm/ then 'arm'
           end

    "https://dl.minio.io/#{dir}/release/#{node['os']}-#{arch}/#{component}"
  end
end
