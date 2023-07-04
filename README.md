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

## Configuration

Before you can start using Uptime-Robot, you need to configure it properly.
Follow these steps to set up the configuration:

1. Open the uptime-robot.yml file in the Uptime-Robot directory.
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

## Usage

Once you have installed and configured Uptime-Robot, you can use it by following
these steps:

1. Open a terminal or command prompt.
2. Navigate to the directory where you placed the Uptime-Robot executable.
3. Run the Uptime-Robot executable:

```bash
./uptime-robot
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
