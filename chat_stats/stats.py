import re
from collections import defaultdict, Counter
import argparse
import os

def process_chat_file(file_path):
    # Initialize message and word count structures
    message_counts = defaultdict(int)
    word_counts = Counter()

    # Word blacklist
    blacklist = set(['media', 'omitted'])

    # Regex patterns
    message_start_pattern = re.compile(r"^\d{2}/\d{2}/\d{4}, \d{2}:\d{2} - .*?: ")
    message_entry_pattern = re.compile(r"^\d{2}/\d{2}/\d{4}, \d{2}:\d{2} - (.*?): (.*)")

    # Open and read the file
    with open(file_path, 'r', encoding='utf-8') as file:
        # Combine lines until a new message is detected
        current_message = ''
        lines_combined = []

        for line in file:
            if message_start_pattern.match(line):
                # If the current line is a new message, append the existing message and start a new one
                if current_message:
                    lines_combined.append(current_message.strip())
                current_message = line.strip()
            else:
                # Otherwise, it's a continuation of a previous message
                current_message += " " + line.strip()

        # Don't forget to add the last message
        if current_message:
            lines_combined.append(current_message.strip())

        # Process each concatenated message line
        for line in lines_combined:
            match = message_entry_pattern.match(line)
            if match:
                user_name = match.group(1)
                message_counts[user_name] += 1

                # Extract words and count them
                word_pattern = re.compile(r"\b[a-zA-Z']+\b")
                message_text = match.group(2)
                words = (word.lower() for word in word_pattern.findall(message_text) if word.lower() not in blacklist)

                # Update word counts
                word_counts.update(words)

    # Total messages sent
    total_messages = sum(message_counts.values())

    # Create a list of tuples, (name, count, percentage)
    stats = [(name, count, (count / total_messages) * 100) for name, count in message_counts.items()]

    def ordinal(n):
        suffix = ['th', 'st', 'nd', 'rd', 'th'][min(n % 10, 4)]
        if 11 <= (n % 100) <= 13:
            suffix = 'th'
        return str(n) + suffix

    # Sort the list by message count (in reverse order so highest is first)
    stats.sort(key=lambda x: x[1], reverse=True)

    # Add numbers and names to the blacklist
    for i in range(10):
        blacklist.add(str(i))
    for name, _, _ in stats:
        blacklist.add(name.lower())

    # Create the result string
    result = []
    format_divider = ' •─────[ {} ]─────•'
    divider        = ' •─────────────────────────•'
    star = ' ->  '
    position = 1

    result.append(format_divider.format("GENERAL"))
    result.append(f"{star}Total Participants : {len(stats)}\n{star}Total Messages : {total_messages}")
    result.append("")
    
    result.append(format_divider.format("LEADERBOARD"))
    for name, count, percentage in stats:
        if position > 1:
            result.append(divider)
        result.append(f"{star} {ordinal(position)} place \n{star} {name}\n{star} {count} messages\n{star} {percentage:.5f}% of messages")
        position += 1

    result.append("")

    result.append(format_divider.format("MOST COMMON WORDS"))
    for word, count in word_counts.most_common(5):
        result.append(f"{star} {word}: {count}")

    result.append(format_divider.format("LEAST COMMON WORDS"))
    for word, count in word_counts.most_common()[:-6:-1]:
        result.append(f"{star} {word}: {count}")

    return "\n".join(result)

if __name__ == '__main__':
    parser = argparse.ArgumentParser(description='Process WhatsApp chat statistics.')
    parser.add_argument('-f', '--file', type=str, required=True, help='Path to the WhatsApp chat text file.')
    args = parser.parse_args()

    output = process_chat_file(args.file)
    print(output)