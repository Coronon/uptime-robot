# Uptime-Robot

Uptime-Robot is a versatile service that allows you to monitor various types of
monitors and push the collected data to Uptime-Kuma. This README will guide you
through the setup and configuration process for using Uptime-Robot effectively.

## Table of Contents

- [Introduction](#introduction)
- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
- [Contributing](#contributing)
- [License](#license)

## Introduction

Uptime-Robot is designed to help you monitor the uptime and performance of your
websites, APIs, servers, and other online services. By configuring different
monitor types, you can collect valuable data and push it to Uptime-Kuma for
further analysis and visualization.

Uptime-Kuma is a powerful open-source tool for monitoring and visualizing uptime
data. It provides a user-friendly interface and a range of features to help you
analyze the collected data and make informed decisions.

## Installation

To use Uptime-Robot, follow these steps:

1. Download the Uptime-Robot executable for your operating system from the
[releases](https://github.com/coronon/uptime-robot/releases) page.

2. Place the executable in a directory of your choice.
3. Install the executable as a system service:

```bash
# You might need to run this command as root/administrator
./uptime-robot -install
```
4. Uptime-Robot is now installed as a system-wide service.

## Configuration

Before you can start using Uptime-Robot, you need to configure it properly.
Uptime-Robot tries to locate its configuration next to the binary.
Follow these steps to set up the configuration:

1. Create a `uptime-robot.yml` file in the Uptime-Robot directory.
2. Customize the configuration options according to your requirements. 

```yaml
# Used to identify this node
node_name: my.hostname

# If you use different uptime-kuma instances you can define multiple hosts
hosts:
  - name: someCoolName
    url: https://status.example.com/api/push/

# These are the monitors that collect data and push it to their hosts
monitors:
  # The name is arbitrary and purely used for logs
  - name: Available disk space
    # Type must be one of the valid types
    type: disk_usage
    # Use the hosts name as defined above
    host: someCoolName
    # This is the unique key Uptime-Kuma gives a push monitor
    # (URL: .../api/push/{YOUR_KEY}?...)
    key: abcdefghij
    # Interval in seconds to run this monitor
    # The time the monitor actually runs does not have an impact on it's
    # scheduling
    interval: 120

    # Arguments specific to the monitor type (if any)
    file_system: C:\
    down_threshold: 95
  - name: Alive ping
    type: alive
    host: someCoolName
    key: zyxwvutsrq
    interval: 60
```

3. Save the configuration file to disk.
4. Restart the Uptime-Robot service.

### Monitor Types

Only configuration options unique to a monitor type will be documented.
For general options, see above. 

#### alive

A simple monitor that periodically sends an **up** status to its host.
This monitor is supposed to signal that a node is still up and connected to the
host.

```yaml
- name: Alive ping
  type: alive
  host: myHostsName
  key: zyxwvutsrq
  interval: 60
```

#### disk_usage

Checks the available disk space in percent and raises a **down** status once the
used space exceeds a configured threshold.

```yaml
- name: Available disk space
  type: disk_usage
  host: myHostsName
  key: zyxwvutsrq
  interval: 60

  # The file system object to check usage of
  # Linux: pathname of any file within a mounted filesystem, e.g. /
  #   When using mounted filesystems the underlying statfs implementation
  #   requires specifying a file WITHIN that filesystem instead of the device.
  #   -> e.g. /mnt/data instead of /dev/sdb
  # Windows: a directory on a disk, e.g. C:
  #   If this parameter is a UNC name, it must include a trailing backslash,
  #   for example, "\\MyServer\MyShare\". This parameter does not have to
  #   specify the root directory on a disk. It accepts any directory on a disk.
  file_path: C:\
  # Usage percentage that will start to trigger a down status
  # actual_usage >= down_threshold ? "DOWN" : "UP"
  down_threshold: 95
```

#### email_ping

Periodically checks whether emails can be sent and received. If either fails
this monitor will raise a **down** status. In order to verify sending and
receiving capabilities, an email is first sent to a reply host such as
[PingPong-Mail](https://github.com/Coronon/pingpong-mail). After sending this
monitor will wait for a specified timeout and check that a reply was received.

```yaml
- name: Email working
  type: email_ping
  host: myHostsName
  key: zyxwvutsrq
  interval: 300

  # The hostname of the SMTP server used to send the initial email 
  smtp_host: smtp.example.com
  # The port of the SMTP server used to send the initial email
  # When using authentication, the most common port to use is 587 for SMTP
  # submission. Other common options are: 25, 2525.
  smtp_port: 587
  # Whether to force TLS encrypted SMTP communications using STARTTLS.
  # The STARTTLS command will always be used if available and cause a down
  # status on error. This option forces a down status if the command is absent.
  smtp_force_tls: true
  # E-Mail address to use in <From:> header (commonly referred to as the sender)
  # This is also the address a reply will be sent to when using PingPong-Mail.
  smtp_sender_address: email-ping@example.com
  # E-Mail address of the initial mails recipient
  # If you don't want to self-host your own auto-reply solution, you can use the
  # address from this example for a free, privacy focused alternative.
  # More information: https://github.com/Coronon/pingpong-mail
  smtp_recipient_address: check@ping-pong.email
  # Username used to authenticate to the SMTP server
  # Most servers will only relay email for authenticated users. If you don't
  # require PLAIN authentication, you may leave this empty.
  smtp_username: email-ping@example.com
  # Password used to authenticate to the SMTP server
  # See explanation above.
  smtp_password: MySuPeRsEcUrEpAsSwOrD

  # The hostname of the IMAP server used to receive the response
  imap_host: imap.example.com
  # The port of the IMAP server used to receive the response
  # Other common options are: 993 (Implicit SSL is not yet supported)
  imap_port: 143
  # Whether to force TLS encrypted IMAP communications using STARTTLS.
  # The STARTTLS command will always be used if available and cause a down
  # status on error. This option forces a down status if the command is absent.
  imap_force_tls: true
  # Username used to authenticate to the IMAP server
  imap_username: email-ping@example.com
  # Password used to authenticate to the IMAP server
  # Optional
  imap_password: MySuPeRsEcUrEpAsSwOrD

  # The subject of the initially sent email
  # The variable `{UUID}` can optionally be used to include a random V4 UUID.
  # It is ok not to include a UUID as old responses are cleaned before each run.
  message_subject: PING ({UUID})
  # The body (text) of the initially sent email
  message_body: This is an automated test message :)
  # Subject expected for email recieved back
  # The variable `{ORIG_SUBJ}` will be replaced with the unaltered subject of
  # the sent email.
  response_subject: "PONG - '{ORIG_SUBJ}'"

  # Time in seconds after which to regard the test as failed if no response was
  # received
  timeout: 180
```

## Usage

Once you have installed and configured Uptime-Robot, you can use it from the
terminal by following these steps:

1. Open a terminal or command prompt.
2. Navigate to the directory where you placed the Uptime-Robot executable.
3. Run the Uptime-Robot executable:

```bash
# The -interactive flag is necessary if uptime-robot is not running under a
# service manager
./uptime-robot -interactive
```

4. Uptime-Robot will start monitoring the specified monitor types and collect
data.
5. The collected data will be pushed to Uptime-Kuma for analysis and
visualization.
6. Access Uptime-Kuma through your preferred web browser and explore the
collected data.

## Contributing

Contributions to Uptime-Robot are welcome! If you encounter any issues or have
suggestions for improvements, please open an issue on the
[GitHub repository](https://github.com/Coronon/uptime-robot/issues).

If you want to contribute to the project, follow these steps:

1. Fork the repository.
2. Create a new branch for your feature or bug fix.
3. Make the necessary changes and commit them.
4. Push your branch to your forked repository.
6. Open a pull request on the main repository and provide a detailed description
of your changes.

## License

Uptime-Robot is open-source software licensed under the
[3-Clause BSD License](https://opensource.org/license/bsd-3-clause/).
See the [LICENSE](https://github.com/Coronon/uptime-robot/blob/master/LICENSE)
file for more details.
