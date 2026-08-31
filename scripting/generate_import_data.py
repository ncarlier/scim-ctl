import json
import random
import sys
import os

DEPARTMENTS = ["Engineering", "Product", "Sales", "Marketing", "HR", "Finance", "Legal", "Operations"]
TITLES = ["Software Engineer", "Product Manager", "Sales Associate", "Marketing Specialist", "HR Partner", "Accountant", "Legal Counsel", "Operations Manager"]
DOMAINS = ["example.com", "test.org", "demo.net", "corp.local"]

def generate_user(index):
    first_name = f"User{index:04d}"
    last_name = f"LastName{index:04d}"
    username = f"user{index:04d}"
    domain = random.choice(DOMAINS)
    
    return {
        "userName": username,
        "name": {
            "givenName": first_name,
            "familyName": last_name
        },
        "emails": [
            {
                "value": f"{username}@{domain}"
            }
        ],
        "phoneNumbers": [
            {
                "value": f"+1-555-{random.randint(1000, 9999)}"
            }
        ],
        "title": random.choice(TITLES),
        "urn:ietf:params:scim:schemas:extension:enterprise:2.0:User": {
            "department": random.choice(DEPARTMENTS),
            "employeeNumber": f"EMP{index:05d}"
        },
        "active": random.choice([True, True, True, False]) # 75% active
    }

def main():
    if len(sys.argv) != 3:
        print("Usage: python3 generate_import_data.py <count> <output_file>")
        sys.exit(1)
        
    try:
        count = int(sys.argv[1])
    except ValueError:
        print("Count must be an integer.")
        sys.exit(1)
        
    output_file = sys.argv[2]
    
    # Ensure directory exists
    os.makedirs(os.path.dirname(output_file), exist_ok=True)
    
    with open(output_file, 'w') as f:
        for i in range(1, count + 1):
            user_data = generate_user(i)
            json_line = json.dumps(user_data)
            f.write(json_line + '\n')
            
    print(f"Successfully generated {count} records in {output_file}")

if __name__ == "__main__":
    main()
